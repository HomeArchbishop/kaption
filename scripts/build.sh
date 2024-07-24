DEV_MODE=false
for arg in "$@"
do
  if [ "$arg" == "--dev" ]; then
    DEV_MODE=true
    echo "[BUILD] Development mode enabled."
    break
  fi
done

if [ "$(uname)" == "Darwin" ]; then
  SUFFIX=""
elif [ "$(expr substr $(uname -s) 1 5)" == "Linux" ]; then
  SUFFIX=""
elif [ "$(expr substr $(uname -s) 1 10)" == "MINGW32_NT" ] || [ "$(expr substr $(uname -s) 1 10)" == "MINGW64_NT" ]; then
  SUFFIX=".exe"
fi

if [ -d "dist" ]; then
  rm -rf ./dist/*
fi

mkdir -p dist

echo "[BUILD] GO building..."
go build -o ./dist/start.exe ./cmd/main

if $DEV_MODE; then
  if [ "$(expr substr $(uname -s) 1 10)" == "MINGW32_NT" ] || [ "$(expr substr $(uname -s) 1 10)" == "MINGW64_NT" ]; then
    # DIST_PATH=$(cd dist && pwd -W | sed 's/\//\\/g')
    # MODEL_PATH=$(cd model && pwd -W | sed 's/\//\\/g')
    # cmd //C mklink /D "$DIST_PATH\\model" "$MODEL_PATH"
    ln -s $(pwd)/model $(pwd)/dist/model
  else
    ln -s $(pwd)/model $(pwd)/dist/model
  fi
else
  cp -r ./model ./dist/
fi

echo "[BUILD] Copying third-party libraries..."
cp -r ./third_party/vosk/src/* ./dist/
cp -r ./third_party/ffmpeg/* ./dist/
