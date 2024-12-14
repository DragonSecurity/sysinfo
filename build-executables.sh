version=`go run . version`

platforms=(
  "darwin/amd64"
  "darwin/arm64"
  "linux/amd64"
  "linux/arm"
  "linux/arm64"
)

for platform in "${platforms[@]}"
do
  platform_split=(${platform//\// })
  GOOS=${platform_split[0]}
  GOARCH=${platform_split[1]}

  os=$GOOS
  if [ $os = "darwin" ]; then
    os="macOS"
  fi

  output_name="sysinfo-${version}-${os}-${GOARCH}"
  if [ $os = "windows" ]; then
    output_name+='.exe'
  fi

  echo "Building release/$output_name..."
  env GOOS=$GOOS GOARCH=$GOARCH go build \
    -ldflags "-X github.com/DragonSecurity/sysinfo/cmd.Version=$version" \
    -o release/$output_name
  if [ $? -ne 0 ]; then
    echo 'An error has occurred! Aborting.'
    exit 1
  fi

  zip_name="sysinfo-${version}-${os}-${GOARCH}"

  # Create checksum for the built file
  checksum_file="${output_name}.sha256"
  echo "Generating checksum for release/$output_name..."
  sha256sum release/$output_name > release/$checksum_file

  pushd release > /dev/null
  if [ $os = "windows" ]; then
    zip $zip_name.zip $output_name $checksum_file
    rm $output_name $checksum_file
  else
    chmod a+x $output_name
    gzip $output_name
    gzip $checksum_file
  fi
  popd > /dev/null
done
