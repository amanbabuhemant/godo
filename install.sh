#!/bin/bash
set -e

BINARY_NAME="godo"
BINARY_URL="https://github.com/biisal/godo/releases/download/v22.6.25/godo"
INSTALL_DIR="$HOME/.local/bin"
COMMNET="# run_godo shortcut"

mkdir -p "$INSTALL_DIR"
echo "Making you more focused on what you want to do..."
curl -sSL "$BINARY_URL" -o "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"
echo "Almost ready..."

LINE='export PATH="$HOME/.local/bin:$PATH"'
if command -v zsh >/dev/null 2>&1; then
    PROFILE="$HOME/.zshrc"
else
    PROFILE="$HOME/.bashrc"
fi

if ! grep -qxF "$LINE" "$PROFILE"; then
    echo "$LINE" >> "$PROFILE"
    echo "🔧 Added ~/.local/bin to PATH in $PROFILE"
fi

if command -v zsh >/dev/null 2>&1; then
	if ! grep -q "$COMMNET" ~/.zshrc; then
	cat >> ~/.zshrc <<EOL
$COMMNET
function run_godo() {
 	 godo
}

zle -N run_godo
bindkey '^g' run_godo
EOL
	fi
fi

if ! grep -q "$COMMNET" ~/.bashrc; then
cat >> ~/.bashrc <<EOL
$COMMNET
bind '"\C-g":"godo\n"'
EOL
fi

echo "Reloading shell..."
exec "$SHELL"


echo "Done. Press Ctrl+g to start.... [laal dil]"
