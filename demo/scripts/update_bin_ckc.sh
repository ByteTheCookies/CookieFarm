#!/bin/bash

# Controlla se il parametro è stato passato
if [ -z "$1" ]; then
  echo "Uso: $0 /path/assoluto/Exploiter"
  exit 1
fi

# Path dell'exploiter
EXPLOITER_PATH="$1"
DIST_PATH="../dist"
BIN_DEST="$EXPLOITER_PATH/cookiefarm/bin"

# Funzione per estrarre e copiare i binari
process_archive() {
  local archive="$1"
  local arch=""
  local os=""
  local bin_name="ckc"

  # Determina l'architettura
  if [[ "$archive" == *"arm64"* ]]; then
    arch="arm64"
  elif [[ "$archive" == *"x86_64"* ]]; then
    arch="x86_64"
  else
    echo "Architettura sconosciuta per $archive"
    return
  fi

  # Determina il sistema operativo
  if [[ "$archive" == *"Darwin"* ]]; then
    os="darwin"
  elif [[ "$archive" == *"Linux"* ]]; then
    os="linux"
  elif [[ "$archive" == *"Windows"* ]]; then
    os="windows"
    bin_name="ckc.exe"
  else
    echo "Sistema operativo sconosciuto per $archive"
    return
  fi

  # Estrae l’archivio in una cartella temporanea
  temp_dir=$(mktemp -d)
  if [[ "$archive" == *.tar.gz ]]; then
    tar -xzf "$archive" -C "$temp_dir"
  elif [[ "$archive" == *.zip ]]; then
    unzip -q "$archive" -d "$temp_dir"
  else
    echo "Formato non supportato: $archive"
    rm -rf "$temp_dir"
    return
  fi

  # Trova il binario dentro la cartella temporanea
  BIN_PATH=$(find "$temp_dir" -type f -name "$bin_name")
  if [ -z "$BIN_PATH" ]; then
    echo "Binario $bin_name non trovato in $archive"
    rm -rf "$temp_dir"
    return
  fi

  # Pulisce la cartella di destinazione
  DEST_DIR="$BIN_DEST/$arch/$os"
  mkdir -p "$DEST_DIR"
  rm -f "$DEST_DIR/"*

  # Copia il binario
  cp "$BIN_PATH" "$DEST_DIR/"

  echo "✔ Copiato $bin_name in $DEST_DIR"

  # Pulisce temporanei
  rm -rf "$temp_dir"
}

# Cicla su tutti gli archivi nella cartella dist
for archive in "$DIST_PATH"/*.tar.gz "$DIST_PATH"/*.zip; do
  [ -e "$archive" ] || continue
  process_archive "$archive"
done
