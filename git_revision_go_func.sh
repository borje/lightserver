GOFILE=currentVersion.go
cat << EOF > $GOFILE
package main

func currentVersion() string {
    return "GIT_HASH"
}
EOF

VERSION=$(git rev-parse --short HEAD)
sed "s/GIT_HASH/$VERSION/g" $GOFILE > $GOFILE.tmp || exit 1
mv $GOFILE.tmp $GOFILE

