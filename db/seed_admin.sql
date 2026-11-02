-- Usuarios del panel de administración (EF-13). No hay registro público: se siembran aquí.
-- Contraseñas solo para desarrollo y pruebas (documentadas en el README); el hash es bcrypt.
--   analista       / Analista2026   (Talento Humano, solo lectura)
--   administrador  / Admin2026      (Comisión de Selección)
INSERT INTO usuarios_admin (usuario, password_hash, rol) VALUES
    ('analista', '$2a$10$EsoS/MzEcP//PVM6C.nEjeZKHH4kwanZCi3QDbCp40ONYMojYm70i', 'analista'),
    ('administrador', '$2a$10$epd9KdxvqX/Yej90RINn8ebdboksSyu0eCeA47fu0BzBeGwustWJe', 'administrador')
ON CONFLICT (usuario) DO NOTHING;
