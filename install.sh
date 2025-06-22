#!/bin/bash
set -e

BINARY_NAME="godo"
BINARY_URL="https://github.com/biisal/godo/releases/download/v22.6.25/godo"
INSTALL_DIR="$HOME/.local/bin"

mkdir -p "$INSTALL_DIR"
echo "Making you more focus on what you want to do..."
curl -sSL "$BINARY_URL" -o "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"
echo "Almost reday....."

COMMNET="# run_godo shortcut"
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

echo "Completed... Now close and open new terminal and Press Ctrl+g ..tq [laal dil]"
