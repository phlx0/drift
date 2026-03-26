# Maintainer: phlx0 <https://github.com/phlx0>
pkgname=drift-git
pkgver=r0.placeholder
pkgrel=1
pkgdesc="Terminal screensaver and idle visualizer with animated scenes (git)"
arch=('x86_64' 'aarch64')
url="https://github.com/phlx0/drift"
license=('MIT')
makedepends=('go' 'git')
provides=('drift')
conflicts=('drift' 'drift-bin')
source=("${pkgname}::git+https://github.com/phlx0/drift.git")
sha256sums=('SKIP')

pkgver() {
    cd "${srcdir}/${pkgname}"
    printf "r%s.%s" "$(git rev-list --count HEAD)" "$(git rev-parse --short HEAD)"
}

build() {
    cd "${srcdir}/${pkgname}"
    export CGO_ENABLED=0
    export GOPATH="${srcdir}/go"
    go build -trimpath \
        -ldflags "-s -w \
            -X main.version=$(git describe --tags --always) \
            -X main.commit=$(git rev-parse --short HEAD) \
            -X main.date=$(git log -1 --format=%cI)" \
        -o drift .
}

package() {
    cd "${srcdir}/${pkgname}"
    install -Dm755 drift "${pkgdir}/usr/bin/drift"
    install -Dm644 LICENSE "${pkgdir}/usr/share/licenses/${pkgname}/LICENSE"
    install -Dm644 README.md "${pkgdir}/usr/share/doc/${pkgname}/README.md"
}
