-- Script único de la base de datos: el DDL de todas las tablas y, al final, los datos semilla
-- de la división política (países, provincias y cantones). Se ejecuta de una vez con
--   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f db/schema.sql
-- y se carga solo al levantar PostgreSQL con docker compose (la primera vez, con el volumen
-- vacío). Cambios al esquema requieren una issue y un PR aprobado.
-- Los usuarios de prueba del panel van aparte, en db/seed_admin.sql, y no se cargan en producción.

CREATE TABLE paises (
    id     SERIAL PRIMARY KEY,
    nombre TEXT NOT NULL UNIQUE
);

CREATE TABLE provincias (
    id      SERIAL PRIMARY KEY,
    pais_id INTEGER NOT NULL REFERENCES paises (id),
    nombre  TEXT NOT NULL,
    UNIQUE (pais_id, nombre)
);

CREATE TABLE cantones (
    id           SERIAL PRIMARY KEY,
    provincia_id INTEGER NOT NULL REFERENCES provincias (id),
    nombre       TEXT NOT NULL,
    UNIQUE (provincia_id, nombre)
);

CREATE TABLE postulantes (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cedula        CHAR(10) NOT NULL UNIQUE,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE postulaciones (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    codigo             TEXT NOT NULL UNIQUE,
    postulante_id      UUID NOT NULL UNIQUE REFERENCES postulantes (id),
    estado             TEXT NOT NULL DEFAULT 'borrador'
                       CHECK (estado IN ('borrador', 'enviada', 'cerrada')),
    cargo              TEXT CHECK (cargo IN ('afiliados', 'jubilados', 'empleador')),
    -- Información personal (copiada del Registro Civil simulado)
    nombres            TEXT NOT NULL,
    apellidos          TEXT NOT NULL,
    genero             TEXT NOT NULL,
    fecha_nacimiento   DATE NOT NULL,
    nacionalidad       TEXT NOT NULL,
    estado_civil       TEXT NOT NULL,
    pareja_nombres     TEXT,
    pareja_cedula      CHAR(10),
    -- Información domiciliaria y de contacto
    pais_id            INTEGER REFERENCES paises (id),
    provincia_id       INTEGER REFERENCES provincias (id),
    canton_id          INTEGER REFERENCES cantones (id),
    provincia_texto    TEXT,
    ciudad_texto       TEXT,
    telefono_celular   TEXT,
    email_secundario   TEXT,
    calle_principal    TEXT,
    numero             TEXT,
    calle_secundaria   TEXT,
    sector             TEXT,
    referencia         TEXT,
    telefono_domicilio TEXT,
    tiene_familiares   BOOLEAN NOT NULL DEFAULT FALSE,
    tiene_publicaciones BOOLEAN NOT NULL DEFAULT FALSE,
    respaldo_organizacion_nombre TEXT,
    respaldo_organizacion_doc_id UUID,
    supero_merito      BOOLEAN NOT NULL DEFAULT FALSE,
    -- Elegible si la pregunta 1 de inhabilidades es Sí y la 2 a la 13 son No (lo calcula el servidor)
    elegible           BOOLEAN NOT NULL DEFAULT FALSE,
    consentimiento_at  TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    enviada_at         TIMESTAMPTZ
);

-- Códigos de postulación (BIESS-AAAA-NNNNNN): únicos y crecientes.
CREATE SEQUENCE IF NOT EXISTS postulacion_codigo_seq;

CREATE TABLE postulacion_familiares (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    postulacion_id    UUID NOT NULL REFERENCES postulaciones (id) ON DELETE CASCADE,
    nombres_apellidos TEXT NOT NULL,
    parentesco        TEXT NOT NULL,
    institucion       TEXT NOT NULL CHECK (institucion IN ('IESS', 'BIESS')),
    area              TEXT NOT NULL,
    cargo             TEXT NOT NULL
);

CREATE TABLE referencias_personales (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    postulacion_id    UUID NOT NULL REFERENCES postulaciones (id) ON DELETE CASCADE,
    nombres_apellidos TEXT NOT NULL,
    relacion          TEXT NOT NULL,
    telefono          TEXT NOT NULL,
    correo            TEXT NOT NULL
);

CREATE TABLE documentos (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    postulacion_id UUID NOT NULL REFERENCES postulaciones (id) ON DELETE CASCADE,
    tipo           TEXT NOT NULL CHECK (tipo IN ('organizacion', 'formacion', 'experiencia', 'publicacion', 'capacitacion')),
    nombre         TEXT NOT NULL,
    tamano_bytes   INTEGER NOT NULL CHECK (tamano_bytes > 0),
    contenido      BYTEA NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE postulaciones
    ADD CONSTRAINT fk_respaldo_doc FOREIGN KEY (respaldo_organizacion_doc_id) REFERENCES documentos (id);

CREATE TABLE formacion_academica (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    postulacion_id  UUID NOT NULL REFERENCES postulaciones (id) ON DELETE CASCADE,
    nivel           TEXT NOT NULL CHECK (nivel IN ('tercer_nivel', 'especialidad', 'maestria', 'phd')),
    institucion     TEXT NOT NULL,
    tipo            TEXT NOT NULL CHECK (tipo IN ('nacional', 'extranjera')),
    titulo          TEXT NOT NULL,
    area            TEXT NOT NULL,
    senescyt_numero TEXT NOT NULL,
    senescyt_fecha  DATE NOT NULL,
    documento_id    UUID REFERENCES documentos (id)
);

CREATE TABLE experiencia_profesional (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    postulacion_id UUID NOT NULL REFERENCES postulaciones (id) ON DELETE CASCADE,
    tipo           TEXT NOT NULL CHECK (tipo IN ('general', 'direccion')),
    cargo          TEXT NOT NULL,
    institucion    TEXT NOT NULL,
    fecha_inicio   DATE NOT NULL,
    fecha_fin      DATE NOT NULL,
    actividades    TEXT NOT NULL,
    descripcion    TEXT NOT NULL CHECK (char_length(descripcion) <= 1000),
    documento_id   UUID NOT NULL REFERENCES documentos (id),
    CHECK (fecha_fin >= fecha_inicio)
);

CREATE TABLE publicaciones (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    postulacion_id UUID NOT NULL REFERENCES postulaciones (id) ON DELETE CASCADE,
    titulo         TEXT NOT NULL,
    tipo           TEXT NOT NULL CHECK (tipo IN ('publicacion', 'investigacion')),
    medio          TEXT NOT NULL,
    fecha          DATE NOT NULL,
    relacionada    BOOLEAN NOT NULL,
    documento_id   UUID REFERENCES documentos (id)
);

CREATE TABLE capacitaciones (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    postulacion_id UUID NOT NULL REFERENCES postulaciones (id) ON DELETE CASCADE,
    tipo           TEXT NOT NULL CHECK (tipo IN ('curso', 'seminario', 'taller')),
    nombre         TEXT NOT NULL,
    institucion    TEXT NOT NULL,
    fecha_inicio   DATE NOT NULL,
    fecha_fin      DATE NOT NULL,
    horas          INTEGER NOT NULL CHECK (horas > 0),
    documento_id   UUID REFERENCES documentos (id),
    CHECK (fecha_fin >= fecha_inicio)
);

CREATE TABLE postulacion_inhabilidades (
    postulacion_id UUID NOT NULL REFERENCES postulaciones (id) ON DELETE CASCADE,
    pregunta       SMALLINT NOT NULL CHECK (pregunta BETWEEN 1 AND 13),
    respuesta      BOOLEAN NOT NULL,
    detalle        JSONB,
    PRIMARY KEY (postulacion_id, pregunta)
);

CREATE TABLE usuarios_admin (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usuario       TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    rol           TEXT NOT NULL CHECK (rol IN ('analista', 'administrador'))
);

CREATE TABLE fases (
    id          SERIAL PRIMARY KEY,
    nombre      TEXT NOT NULL UNIQUE,
    orden       SMALLINT NOT NULL UNIQUE,
    publicada   BOOLEAN NOT NULL DEFAULT FALSE,
    habilita_formulario BOOLEAN NOT NULL DEFAULT FALSE,
    -- Periodo de la fase (opcional: sin fecha no hay límite por ese lado). Para la fase
    -- Postulación es el periodo en que se puede crear, guardar y enviar la postulación.
    fecha_inicio TIMESTAMPTZ,
    fecha_fin    TIMESTAMPTZ,
    CHECK (fecha_inicio IS NULL OR fecha_fin IS NULL OR fecha_fin >= fecha_inicio)
);

-- Fases del concurso. Postulación arranca publicada y con el formulario habilitado, sin
-- fechas, para que el sistema funcione al levantarlo; el administrador fija el periodo real
-- desde el panel (/admin/fases).
INSERT INTO fases (nombre, orden, publicada, habilita_formulario) VALUES
    ('Postulación', 1, TRUE, TRUE),
    ('Méritos', 2, FALSE, FALSE),
    ('Evaluación', 3, FALSE, FALSE);

CREATE TABLE eval_config (
    id                SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    duracion_minutos  INTEGER NOT NULL DEFAULT 30 CHECK (duracion_minutos > 0),
    num_preguntas     INTEGER NOT NULL DEFAULT 10 CHECK (num_preguntas > 0)
);

CREATE TABLE eval_preguntas (
    id                 SERIAL PRIMARY KEY,
    enunciado          TEXT NOT NULL,
    opciones           JSONB NOT NULL,
    respuesta_correcta SMALLINT NOT NULL
);

CREATE TABLE eval_intentos (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    postulacion_id UUID NOT NULL UNIQUE REFERENCES postulaciones (id),
    preguntas     JSONB NOT NULL,
    inicio        TIMESTAMPTZ NOT NULL DEFAULT now(),
    cierre        TIMESTAMPTZ NOT NULL,
    finalizado_at TIMESTAMPTZ,
    nota          NUMERIC(5,2)
);

CREATE TABLE eval_respuestas (
    intento_id  UUID NOT NULL REFERENCES eval_intentos (id) ON DELETE CASCADE,
    pregunta_id INTEGER NOT NULL REFERENCES eval_preguntas (id),
    opcion      SMALLINT NOT NULL,
    PRIMARY KEY (intento_id, pregunta_id)
);

CREATE INDEX idx_provincias_pais ON provincias (pais_id);
CREATE INDEX idx_cantones_provincia ON cantones (provincia_id);
CREATE INDEX idx_familiares_postulacion ON postulacion_familiares (postulacion_id);
CREATE INDEX idx_referencias_postulacion ON referencias_personales (postulacion_id);
CREATE INDEX idx_documentos_postulacion ON documentos (postulacion_id);
CREATE INDEX idx_formacion_postulacion ON formacion_academica (postulacion_id);
CREATE INDEX idx_experiencia_postulacion ON experiencia_profesional (postulacion_id);
CREATE INDEX idx_postulaciones_cargo_estado ON postulaciones (cargo, estado);

-- =====================================================================================
-- Datos semilla de la división política (catálogos geográficos)
--
-- Tomados de los ítems 1.2.1, 1.2.2 y 1.2.3 de docs/requerimiento-biess.md: 194 países,
-- las 24 provincias del Ecuador y sus 199 cantones. Los ids se resuelven por nombre, no se
-- escriben a mano. Van en este mismo script, después del DDL, para que un solo archivo cree
-- todas las tablas y deje cargados los catálogos.
-- =====================================================================================

INSERT INTO paises (nombre) VALUES
    ('Afganistán'),
    ('Albania'),
    ('Alemania'),
    ('Andorra'),
    ('Angola'),
    ('Antigua y Barbuda'),
    ('Arabia Saudita'),
    ('Argelia'),
    ('Argentina'),
    ('Armenia'),
    ('Australia'),
    ('Austria'),
    ('Azerbaiyán'),
    ('Bahamas'),
    ('Bangladés'),
    ('Barbados'),
    ('Baréin'),
    ('Bélgica'),
    ('Belice'),
    ('Benín'),
    ('Bielorrusia'),
    ('Birmania'),
    ('Bolivia'),
    ('Bosnia y Herzegovina'),
    ('Botsuana'),
    ('Brasil'),
    ('Brunéi'),
    ('Bulgaria'),
    ('Burkina Faso'),
    ('Burundi'),
    ('Bután'),
    ('Cabo Verde'),
    ('Camboya'),
    ('Camerún'),
    ('Canadá'),
    ('Catar'),
    ('Chad'),
    ('Chile'),
    ('China'),
    ('Chipre'),
    ('Ciudad del Vaticano'),
    ('Colombia'),
    ('Comoras'),
    ('Corea del Norte'),
    ('Corea del Sur'),
    ('Costa de Marfil'),
    ('Costa Rica'),
    ('Croacia'),
    ('Cuba'),
    ('Dinamarca'),
    ('Dominica'),
    ('Ecuador'),
    ('Egipto'),
    ('El Salvador'),
    ('Emiratos Árabes Unidos'),
    ('Eritrea'),
    ('Eslovaquia'),
    ('Eslovenia'),
    ('España'),
    ('Estados Unidos'),
    ('Estonia'),
    ('Etiopía'),
    ('Filipinas'),
    ('Finlandia'),
    ('Fiyi'),
    ('Francia'),
    ('Gabón'),
    ('Gambia'),
    ('Georgia'),
    ('Ghana'),
    ('Granada'),
    ('Grecia'),
    ('Guatemala'),
    ('Guyana'),
    ('Guinea'),
    ('Guinea-Bisáu'),
    ('Guinea Ecuatorial'),
    ('Haití'),
    ('Honduras'),
    ('Hungría'),
    ('India'),
    ('Indonesia'),
    ('Irak'),
    ('Irán'),
    ('Irlanda'),
    ('Islandia'),
    ('Islas Marshall'),
    ('Islas Salomón'),
    ('Israel'),
    ('Italia'),
    ('Jamaica'),
    ('Japón'),
    ('Jordania'),
    ('Kazajistán'),
    ('Kenia'),
    ('Kirguistán'),
    ('Kiribati'),
    ('Kuwait'),
    ('Laos'),
    ('Lesoto'),
    ('Letonia'),
    ('Líbano'),
    ('Liberia'),
    ('Libia'),
    ('Liechtenstein'),
    ('Lituania'),
    ('Luxemburgo'),
    ('Macedonia del Norte'),
    ('Madagascar'),
    ('Malasia'),
    ('Malaui'),
    ('Maldivas'),
    ('Malí'),
    ('Malta'),
    ('Marruecos'),
    ('Mauricio'),
    ('Mauritania'),
    ('México'),
    ('Micronesia'),
    ('Moldavia'),
    ('Mónaco'),
    ('Mongolia'),
    ('Montenegro'),
    ('Mozambique'),
    ('Namibia'),
    ('Nauru'),
    ('Nepal'),
    ('Nicaragua'),
    ('Níger'),
    ('Nigeria'),
    ('Noruega'),
    ('Nueva Zelanda'),
    ('Omán'),
    ('Países Bajos'),
    ('Pakistán'),
    ('Palaos'),
    ('Panamá'),
    ('Papúa Nueva Guinea'),
    ('Paraguay'),
    ('Perú'),
    ('Polonia'),
    ('Portugal'),
    ('Reino Unido'),
    ('República Centroafricana'),
    ('República Checa'),
    ('República del Congo'),
    ('República Democrática del Congo'),
    ('República Dominicana'),
    ('Ruanda'),
    ('Rumanía'),
    ('Rusia'),
    ('Samoa'),
    ('San Cristóbal y Nieves'),
    ('San Marino'),
    ('San Vicente y las Granadinas'),
    ('Santa Lucía'),
    ('Santo Tomé y Príncipe'),
    ('Senegal'),
    ('Serbia'),
    ('Seychelles'),
    ('Sierra Leona'),
    ('Singapur'),
    ('Siria'),
    ('Somalia'),
    ('Sri Lanka'),
    ('Suazilandia'),
    ('Sudáfrica'),
    ('Sudán'),
    ('Sudán del Sur'),
    ('Suecia'),
    ('Suiza'),
    ('Surinam'),
    ('Tailandia'),
    ('Tanzania'),
    ('Tayikistán'),
    ('Timor Oriental'),
    ('Togo'),
    ('Tonga'),
    ('Trinidad y Tobago'),
    ('Túnez'),
    ('Turkmenistán'),
    ('Turquía'),
    ('Tuvalu'),
    ('Ucrania'),
    ('Uganda'),
    ('Uruguay'),
    ('Uzbekistán'),
    ('Vanuatu'),
    ('Venezuela'),
    ('Vietnam'),
    ('Yemen'),
    ('Yibuti'),
    ('Zambia'),
    ('Zimbabue');

INSERT INTO provincias (pais_id, nombre)
SELECT (SELECT id FROM paises WHERE nombre = 'Ecuador'), v.nombre
FROM (VALUES
    ('Azuay'),
    ('Bolívar'),
    ('Cañar'),
    ('Carchi'),
    ('Chimborazo'),
    ('Cotopaxi'),
    ('El Oro'),
    ('Esmeraldas'),
    ('Galápagos'),
    ('Guayas'),
    ('Imbabura'),
    ('Loja'),
    ('Los Ríos'),
    ('Manabí'),
    ('Morona Santiago'),
    ('Napo'),
    ('Orellana'),
    ('Pastaza'),
    ('Pichincha'),
    ('Santa Elena'),
    ('Santo Domingo de los Tsáchilas'),
    ('Sucumbíos'),
    ('Tungurahua'),
    ('Zamora Chinchipe')
) AS v (nombre);

INSERT INTO cantones (provincia_id, nombre)
SELECT p.id, v.canton
FROM (VALUES
    ('Azuay', 'Cuenca'),
    ('Azuay', 'Gualaceo'),
    ('Azuay', 'Chordeleg'),
    ('Azuay', 'Sígsig'),
    ('Azuay', 'Girón'),
    ('Azuay', 'Santa Isabel'),
    ('Azuay', 'Pucará'),
    ('Azuay', 'Camilo Ponce Enríquez'),
    ('Bolívar', 'Guaranda'),
    ('Bolívar', 'Chillanes'),
    ('Bolívar', 'Chimbo'),
    ('Bolívar', 'Echeandía'),
    ('Bolívar', 'San Miguel'),
    ('Bolívar', 'Caluma'),
    ('Bolívar', 'Las Naves'),
    ('Cañar', 'Azogues'),
    ('Cañar', 'Biblián'),
    ('Cañar', 'Cañar'),
    ('Cañar', 'Déleg'),
    ('Cañar', 'El Tambo'),
    ('Cañar', 'Suscal'),
    ('Cañar', 'La Troncal'),
    ('Carchi', 'Tulcán'),
    ('Carchi', 'Bolívar'),
    ('Carchi', 'Espejo'),
    ('Carchi', 'Mira'),
    ('Carchi', 'Montúfar'),
    ('Carchi', 'San Pedro de Huaca'),
    ('Chimborazo', 'Riobamba'),
    ('Chimborazo', 'Alausí'),
    ('Chimborazo', 'Colta'),
    ('Chimborazo', 'Chunchi'),
    ('Chimborazo', 'Guamote'),
    ('Chimborazo', 'Guano'),
    ('Chimborazo', 'Pallatanga'),
    ('Chimborazo', 'Penipe'),
    ('Chimborazo', 'Cumandá'),
    ('Cotopaxi', 'Latacunga'),
    ('Cotopaxi', 'La Maná'),
    ('Cotopaxi', 'Pangua'),
    ('Cotopaxi', 'Pujilí'),
    ('Cotopaxi', 'Salcedo'),
    ('Cotopaxi', 'Saquisilí'),
    ('Cotopaxi', 'Sigchos'),
    ('El Oro', 'Machala'),
    ('El Oro', 'Arenillas'),
    ('El Oro', 'Atahualpa'),
    ('El Oro', 'Balsas'),
    ('El Oro', 'Chilla'),
    ('El Oro', 'El Guabo'),
    ('El Oro', 'Huaquillas'),
    ('El Oro', 'Las Lajas'),
    ('El Oro', 'Marcabelí'),
    ('El Oro', 'Pasaje'),
    ('El Oro', 'Piñas'),
    ('El Oro', 'Portovelo'),
    ('El Oro', 'Santa Rosa'),
    ('El Oro', 'Zaruma'),
    ('Esmeraldas', 'Esmeraldas'),
    ('Esmeraldas', 'Atacames'),
    ('Esmeraldas', 'Eloy Alfaro'),
    ('Esmeraldas', 'Muisne'),
    ('Esmeraldas', 'Quinindé'),
    ('Esmeraldas', 'Rioverde'),
    ('Esmeraldas', 'San Lorenzo'),
    ('Galápagos', 'San Cristóbal'),
    ('Galápagos', 'Isabela'),
    ('Galápagos', 'Santa Cruz'),
    ('Guayas', 'Guayaquil'),
    ('Guayas', 'Durán'),
    ('Guayas', 'Milagro'),
    ('Guayas', 'Samborondón'),
    ('Guayas', 'Daule'),
    ('Guayas', 'Playas'),
    ('Guayas', 'Balzar'),
    ('Guayas', 'Colimes'),
    ('Guayas', 'El Empalme'),
    ('Guayas', 'El Triunfo'),
    ('Guayas', 'Isidro Ayora'),
    ('Guayas', 'Lomas de Sargentillo'),
    ('Guayas', 'Naranjal'),
    ('Guayas', 'Naranjito'),
    ('Guayas', 'Nobol'),
    ('Guayas', 'Pedro Carbo'),
    ('Guayas', 'Salitre'),
    ('Guayas', 'Simón Bolívar'),
    ('Guayas', 'Yaguachi'),
    ('Imbabura', 'Ibarra'),
    ('Imbabura', 'Antonio Ante'),
    ('Imbabura', 'Cotacachi'),
    ('Imbabura', 'Otavalo'),
    ('Imbabura', 'Pimampiro'),
    ('Imbabura', 'San Miguel de Urcuquí'),
    ('Loja', 'Loja'),
    ('Loja', 'Calvas'),
    ('Loja', 'Catamayo'),
    ('Loja', 'Celica'),
    ('Loja', 'Chaguarpamba'),
    ('Loja', 'Espíndola'),
    ('Loja', 'Gonzanamá'),
    ('Loja', 'Macará'),
    ('Loja', 'Paltas'),
    ('Loja', 'Puyango'),
    ('Loja', 'Saraguro'),
    ('Loja', 'Sozoranga'),
    ('Loja', 'Zapotillo'),
    ('Los Ríos', 'Babahoyo'),
    ('Los Ríos', 'Baba'),
    ('Los Ríos', 'Buena Fe'),
    ('Los Ríos', 'Mocache'),
    ('Los Ríos', 'Montalvo'),
    ('Los Ríos', 'Palenque'),
    ('Los Ríos', 'Puebloviejo'),
    ('Los Ríos', 'Quevedo'),
    ('Los Ríos', 'Urdaneta'),
    ('Los Ríos', 'Ventanas'),
    ('Los Ríos', 'Vinces'),
    ('Los Ríos', 'Valencia'),
    ('Manabí', 'Portoviejo'),
    ('Manabí', 'Manta'),
    ('Manabí', 'Chone'),
    ('Manabí', 'Bolívar'),
    ('Manabí', 'El Carmen'),
    ('Manabí', 'Flavio Alfaro'),
    ('Manabí', 'Jipijapa'),
    ('Manabí', 'Junín'),
    ('Manabí', 'Montecristi'),
    ('Manabí', 'Paján'),
    ('Manabí', 'Pichincha'),
    ('Manabí', 'Rocafuerte'),
    ('Manabí', 'Santa Ana'),
    ('Manabí', 'Sucre'),
    ('Manabí', 'Tosagua'),
    ('Manabí', '24 de Mayo'),
    ('Manabí', 'Pedernales'),
    ('Manabí', 'Olmedo'),
    ('Manabí', 'Puerto López'),
    ('Manabí', 'Jama'),
    ('Manabí', 'Jaramijó'),
    ('Manabí', 'San Vicente'),
    ('Morona Santiago', 'Morona'),
    ('Morona Santiago', 'Gualaquiza'),
    ('Morona Santiago', 'Limón Indanza'),
    ('Morona Santiago', 'Palora'),
    ('Morona Santiago', 'San Juan Bosco'),
    ('Morona Santiago', 'Santiago'),
    ('Morona Santiago', 'Sucúa'),
    ('Morona Santiago', 'Taisha'),
    ('Morona Santiago', 'Tiwintza'),
    ('Napo', 'Tena'),
    ('Napo', 'Archidona'),
    ('Napo', 'Carlos Julio Arosemena Tola'),
    ('Napo', 'El Chaco'),
    ('Napo', 'Quijos'),
    ('Orellana', 'Francisco de Orellana'),
    ('Orellana', 'Aguarico'),
    ('Orellana', 'La Joya de los Sachas'),
    ('Orellana', 'Loreto'),
    ('Pastaza', 'Pastaza'),
    ('Pastaza', 'Arajuno'),
    ('Pastaza', 'Mera'),
    ('Pastaza', 'Santa Clara'),
    ('Pichincha', 'Quito'),
    ('Pichincha', 'Cayambe'),
    ('Pichincha', 'Mejía'),
    ('Pichincha', 'Pedro Moncayo'),
    ('Pichincha', 'Pedro Vicente Maldonado'),
    ('Pichincha', 'Puerto Quito'),
    ('Pichincha', 'Rumiñahui'),
    ('Pichincha', 'San Miguel de los Bancos'),
    ('Santa Elena', 'Santa Elena'),
    ('Santa Elena', 'La Libertad'),
    ('Santo Domingo de los Tsáchilas', 'La Concordia'),
    ('Santo Domingo de los Tsáchilas', 'Santo Domingo'),
    ('Sucumbíos', 'Cascales'),
    ('Sucumbíos', 'Cuyabeno'),
    ('Sucumbíos', 'Gonzalo Pizarro'),
    ('Sucumbíos', 'Lago Agrio'),
    ('Sucumbíos', 'Putumayo'),
    ('Sucumbíos', 'Shushufindi'),
    ('Sucumbíos', 'Sucumbíos'),
    ('Tungurahua', 'Ambato'),
    ('Tungurahua', 'Baños de Agua Santa'),
    ('Tungurahua', 'Cevallos'),
    ('Tungurahua', 'Mocha'),
    ('Tungurahua', 'Patate'),
    ('Tungurahua', 'Pelileo'),
    ('Tungurahua', 'Píllaro'),
    ('Tungurahua', 'Quero'),
    ('Tungurahua', 'Tisaleo'),
    ('Zamora Chinchipe', 'Centinela del Cóndor'),
    ('Zamora Chinchipe', 'Chinchipe'),
    ('Zamora Chinchipe', 'El Pangui'),
    ('Zamora Chinchipe', 'Nangaritza'),
    ('Zamora Chinchipe', 'Palanda'),
    ('Zamora Chinchipe', 'Paquisha'),
    ('Zamora Chinchipe', 'Yacuambi'),
    ('Zamora Chinchipe', 'Yantzaza'),
    ('Zamora Chinchipe', 'Zamora')
) AS v (provincia, canton)
JOIN provincias p
  ON p.nombre = v.provincia
 AND p.pais_id = (SELECT id FROM paises WHERE nombre = 'Ecuador');
