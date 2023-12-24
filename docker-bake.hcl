variable "VERSION" {
  default = "0.1.1"
}

group "default" {
  targets = ["default"]
}

target "default" {
  platforms = ["linux/amd64", "linux/arm64"]
  tags = ["quay.io/seiferma/mediathek2rss:${VERSION}", "quay.io/seiferma/mediathek2rss:latest"]
}
