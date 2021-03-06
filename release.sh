#!/bin/bash

function usage {
    echo -e "colorart release script\n"
    echo "Usage:"
    echo "  $0 version"
    exit 1
}

version=$1
if [ -z "$version" ]; then
    usage
fi

if  [ ! -d "bin" ]; then
    mkdir bin
fi



function xc {
    echo ">>> Cross compiling colorart"
    cd bin/
    GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=${version}" -o colorart-${version}-linux-amd64
    GOOS=linux GOARCH=386 go build -ldflags "-X main.version=${version}" -o colorart-${version}-linux-i386
    GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=${version}" -o colorart-${version}-darwin-amd64
    cd ../
}

function deb {
    arches="i386 amd64"
    for arch in $arches; do
        echo -e "\n>>> Creating debian package for ${arch}"
        fpm \
            -f \
            -s dir \
            -t deb \
            --vendor "DeLaTech" \
            --name   "colorart" \
            --description "An utility that extracts colors from an image (similiar to iTunes 11+)" \
            --version $version \
            -a $arch \
            -p ./bin/colorart-${version}-${arch}.deb \
            ./bin/colorart-${version}-linux-${arch}=/usr/bin/colorart
    done
}

function osx {
    echo -e "\n>>> Creating osx package"
    fpm \
        -f \
        -s dir \
        -t tar \
        --name   "colorart" \
        -p ./bin/colorart-darwin-${version}.tar \
        ./bin/colorart-${version}-darwin-amd64=/usr/bin/colorart
}

function publish_debian {
echo -e ">>> Publishing debian packages"
    aptly repo create delatech
    aptly repo add delatech bin/colorart-${version}-i386.deb
    aptly repo add delatech bin/colorart-${version}-amd64.deb
    aptly snapshot create delatech-colorart-${version} from repo delatech


    # for first tie use
    # aptly publish -distribution=squeeze snapshot delatech-colorart-${version} s3:apt.delatech.net:
    aptly publish switch squeeze s3:apt.delatech.net: delatech-colorart-${version}

}

function publish_homebrew {
    echo -e "\n>>> Publishing osx package"
    gzip -f ./bin/colorart-darwin-${version}.tar

    sha1sum=`sha1sum ./bin/colorart-darwin-${version}.tar.gz | awk '{print $1}'`
    aws s3 cp bin/colorart-darwin-${version}.tar.gz s3://release.delatech.net/colorart/colorart-${version}.tar.gz --acl=public-read

    cat <<EOF > $DELATECH_BREWTAP/Formula/colorart.rb
#encoding: utf-8

require 'formula'

class Colorart < Formula
    homepage 'https://github.com/delatech/colorart'
    version '${version}'

    url 'http://release.delatech.net.s3-website-eu-west-1.amazonaws.com/colorart/colorart-${version}.tar.gz'
    sha1 '${sha1sum}'

    depends_on :arch => :intel

    def install
        bin.install 'bin/colorart'
    end
end
EOF
    cd $DELATECH_BREWTAP
    git add Formula/colorart.rb
    git ci -m"Update colorart to v${version}"
    git push origin master
}

export AWS_ACCESS_KEY_ID=$AWS_DELATECH_S3_APT_KEY
export AWS_SECRET_ACCESS_KEY=$AWS_DELATECH_S3_APT_SECRET

xc
deb
publish_debian
osx
publish_homebrew
