-- Esquema base (borrador). Cambios al esquema requieren una issue y un PR aprobado.
-- Se carga automáticamente al levantar PostgreSQL con docker-compose.

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
    habilita_formulario BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE eval_config (
    id                SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    inicia_en         TIMESTAMPTZ,
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
CREATE INDEX idx_documentos_postulacion ON documentos (postulacion_id);
CREATE INDEX idx_formacion_postulacion ON formacion_academica (postulacion_id);
CREATE INDEX idx_experiencia_postulacion ON experiencia_profesional (postulacion_id);
CREATE INDEX idx_postulaciones_cargo_estado ON postulaciones (cargo, estado);
