# garbo

`garbo` is a minimalistic macOS app to view archives: zip, tar, bzip2, xz, and
others.

`garbo` uses the [mholt/archiver](https://github.com/mholt/archiver) library for
archive file handling.

`garbo` uses the [fyne.io](https://fyne.io) cross-platform UI toolkit for its GUI.

`garbo` is still in early stages of development. Please create a new issue or send
a pull request to help improve it!

The scope of this project is kept deliberately narrow:
* browsing of existing archives
  - open an archive from the macOS Finder or Terminal
* extracting of single or multiple files from an archive
  - no focus on creating archives or added / changing / removing files from existing archives
* macOS as the only targeted platform
  - no focus on Windows, Linux, or mobile platforms

The name **garbo** may or may not be derived from **G**o **AR**chive **B**r**O**wser.

## Alternative Projects

[7-Zip](https://www.7-zip.org/) does not have a macOS GUI application.

[keka](https://www.keka.io/en/) does not have an "archive browser" interface, only extraction capabilities.

[BetterZip](https://macitbetter.com/) is a great GUI, but not open source or free software. 

## Build

`fyne package -name garbo -os darwin`

## Contributing

`garbo` welcomes contributions from anyone! Unlike many other projects we are
happy to accept cosmetic contributions and small contributions, in addition to
large feature requests and changes.

## License

`garbo` is made available under the MIT License. For more details, see the
`LICENSE` file in the repository.

The included application icon is 
from https://publicdomainvectors.org/en/free-clipart/History-icon/45339.html 
via https://openclipart.org/detail/248168/history-icon
by "thewizardplusplus" under Public Domain license.

## Authors

`garbo` was created by Thomas Kriechbaumer, and is maintained by the community.
