#\!/bin/bash
# Script temporaire pour commenter les imports ent problématiques
files="internal/modules/alertes/repository.go internal/modules/pv/repository.go internal/modules/commissariat/repository.go internal/modules/admin/repository.go internal/modules/controles/controller.go internal/modules/controles/repository.go"

for file in $files; do
  if [ -f "$file" ]; then
    echo "Processing $file"
    sed -i.bak "s|^[[:space:]]*\"police-trafic-api-frontend-aligned/ent/|// &|g" "$file"
    sed -i.bak2 "s|^[[:space:]]*\".*ent/[a-zA-Z]*\"|// &|g" "$file"
  fi
done
