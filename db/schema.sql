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
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    enviada_at         TIMESTAMPTZ
);

CREATE TABLE postulacion_familiares (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    postulacion_id    UUID NOT NULL REFERENCES postulaciones (id) ON DELETE CASCADE,
    nombres_apellidos TEXT NOT NULL,
    parentesco        TEXT NOT NULL,
    institucion       TEXT NOT NULL CHECK (institucion IN ('IESS', 'BIESS')),
    area              TEXT NOT NULL,
    cargo             TEXT NOT NULL
);

CREATE INDEX idx_provincias_pais ON provincias (pais_id);
CREATE INDEX idx_cantones_provincia ON cantones (provincia_id);
CREATE INDEX idx_familiares_postulacion ON postulacion_familiares (postulacion_id);
