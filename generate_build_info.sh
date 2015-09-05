GOFILE=buildinfo.go
cat << EOF > $GOFILE
package main

func currentVersion() string {
    return "GIT_HASH"
}

func buildTime() string {
    return "BUILD_TIME"
}
EOF

VERSION=$(git rev-parse --short HEAD)
BUILDTIME=$(stat -f "%Sm" lightserver)
sed -e "s/GIT_HASH/$VERSION/g" -e "s/BUILD_TIME/$BUILDTIME/g" $GOFILE > $GOFILE.tmp || exit 1
mv $GOFILE.tmp $GOFILE

