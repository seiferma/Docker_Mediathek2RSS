# Mediathek2RSS
[![](https://github.com/seiferma/Docker_Mediathek2RSS/actions/workflows/docker-publish.yml/badge.svg?branch=main)](https://github.com/seiferma/Docker_Mediathek2RSS/actions?query=branch%3Amain+)
[![](https://img.shields.io/github/issues/seiferma/Docker_Mediathek2RSS.svg)](https://github.com/seiferma/Docker_Mediathek2RSS/issues)
[![](https://img.shields.io/github/license/seiferma/Docker_Mediathek2RSS.svg)](https://github.com/seiferma/Docker_Mediathek2RSS/blob/main/LICENSE)

The web service provides RSS podcast feeds for shows from German public television. As of now the web service considers
* [ARD Mediathek](https://www.ardmediathek.de)
* [ZDF Mediathek](https://www.zdf.de)

## Usage
The docker image is available under `quay.io/seiferma/mediathek2rss:latest` (see the [registry page](https://quay.io/repository/seiferma/mediathek2rss?tab=tags) for available versions). When running a container, the web service listenes to requests on port `8080`.

Different channels usually have different ways to identify shows. Have a look at the following paragraphs for detailed information about this.

All services support asking for a preferred quality by giving the expected media width in pixels. The width is passed as query parameter by appending `?width={n}` to the URL. For instance, by specifying `720`, you request a HD ready video stream. The web service tries to meet this request as close as possible. It is possible to filter episodes by its length. By appending the query parameter `?minLength={n}` to the URL, all episodes that have less than `n` seconds will not be part of the RSS feed.

To avoid spamming the API of television channels, feeds are only regenerated every 5 minutes on request.

### ARD Shows
The RSS feed for ARD shows is available via `/ard/show/{showID}`. The show ID is a alphanumeric string that you can collect from the show's URL in the mediathek. For instance, `Y3JpZDovL2Z1bmsubmV0LzEwMzE` is the show id for the show `Walulis`, which has the URL `https://www.ardmediathek.de/ard/sendung/walulis/Y3JpZDovL2Z1bmsubmV0LzEwMzE/`. 

### ZDF Shows
The RSS feed for ZDF shows is available via `/zdf/show/byPath/{showPath}`. The show path is a substring of the URL to the show. For instance, `comedy/zdf-magazin-royale` is the show path for the show `ZDF Magazin Royale`, which has the URL `https://www.zdf.de/comedy/zdf-magazin-royale`.
