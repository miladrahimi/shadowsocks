VER=1.4.0
URL=https://github.com/Jigsaw-Code/outline-ss-server/releases/download

OS=$(uname)
ARCH=$(uname -m)
BASE=$( dirname -- "$0"; )

echo "Running outline-ss-server.sh"

if [ "$OS-$ARCH" == "Darwin-arm64" ]; then
  TARGET=outline-ss-server_1.4.0_macos_arm64
  DIR="${BASE}/outline-macos-arm64"
  if [ ! -f "${DIR}/outline-ss-server" ]; then
    mkdir -p "${DIR}"
    wget -qc "${URL}/v${VER}/${TARGET}.tar.gz" -O - | tar -xzC "${DIR}"
  fi
else
  TARGET=outline-ss-server_1.4.0_linux_x86_64
  DIR="${BASE}/outline-linux-x86_64"
  if [ ! -f "${DIR}/outline-ss-server" ]; then
    mkdir -p "${DIR}"
    wget -qc "${URL}/v${VER}/${TARGET}.tar.gz" -O - | tar -xzC "${DIR}"
  fi
fi