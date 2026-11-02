#!/usr/bin/env bash
# Respaldo de PostgreSQL con pg_dump, comprimido y con rotación.
#
# Uso (desde cualquier carpeta, en el servidor):
#   /ruta/al/repo/scripts/backup.sh
#
# Diario con cron (por ejemplo a las 02:30), como el usuario que corre Docker:
#   30 2 * * * /ruta/al/repo/scripts/backup.sh >> $HOME/backups/biess/backup.log 2>&1
#
# Variables opcionales:
#   BACKUP_DIR            carpeta de destino   (por defecto: $HOME/backups/biess)
#   BACKUP_RETENTION_DAYS días que se conservan (por defecto: 7)
#
# Restaurar un respaldo en una base vacía:
#   gunzip -c biess-AAAAmmdd-HHMMSS.sql.gz | docker compose -f docker-compose.prod.yml exec -T postgres sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"'
set -euo pipefail

# Los respaldos tienen datos personales (cédulas, correos y los PDF de los postulantes):
# la carpeta y cada archivo se crean solo para el dueño desde el principio, no recién al
# terminar. Sin esto, el archivo parcial y la carpeta nacen legibles por otros usuarios.
umask 077

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_DIR="${BACKUP_DIR:-$HOME/backups/biess}"
RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-7}"

case "$RETENTION_DAYS" in
  ''|*[!0-9]*) echo "BACKUP_RETENTION_DAYS debe ser un número entero de días" >&2; exit 1 ;;
esac

# Se trabaja desde la raíz del repo para que docker compose encuentre su archivo y su .env.
cd "$REPO_DIR"

mkdir -p "$BACKUP_DIR"
destino="$BACKUP_DIR/biess-$(date +%Y%m%d-%H%M%S).sql.gz"
parcial="$destino.parcial"
trap 'rm -f "$parcial"' EXIT

# pg_dump corre dentro del contenedor, con el usuario y la base de su propio entorno.
# Se escribe a un archivo parcial y solo se renombra si todo salió bien: un respaldo a
# medias nunca queda con el nombre definitivo.
docker compose -f docker-compose.prod.yml exec -T postgres \
  sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" --no-owner' | gzip > "$parcial"

# Un respaldo vacío (solo el encabezado de gzip) significa que pg_dump no produjo nada.
if [ "$(gzip -dc "$parcial" | wc -c)" -eq 0 ]; then
  echo "El respaldo salió vacío; no se guarda" >&2
  exit 1
fi
mv "$parcial" "$destino"
chmod 600 "$destino"
echo "Respaldo creado: $destino"

# Rotación: borra los respaldos con más días que los configurados.
find "$BACKUP_DIR" -maxdepth 1 -name 'biess-*.sql.gz' -type f -mtime +"$RETENTION_DAYS" -print -delete
