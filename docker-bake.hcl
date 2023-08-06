variable "VERSION" {
  default = "0.1.1"
}

group "default" {
  targets = ["default"]
}

target "default" {
  tags = ["quay.io/seiferma/mediathek2rss:${VERSION}", "quay.io/seiferma/mediathek2rss:latest"]
}
