# Hub tool
Docker cli tool to play with Docker Hub.

## Install

The release binaries for all platforms are in [GitHub
Releases](https://github.com/docker/hub-tool/releases). Unfortunately, we do not
support ARM right now.

### Mac OSX
Download the binary from GitHub releases and then run `chmod +x ./hub-tool_darwin_amd64`.

After this, there are a few more hoops to jump through in order for it to
properly install. [Here is a screen recording showing all of the steps if you prefer that](https://github.com/ingshtrom/hub-tool/blob/osx-install-screencast/media/osx_install.mp4).

1. Attempt to run the binary `./hub-tool_darwin_amd64` and choose the "Cancel"
   button in the security prompt
   <br />
   <img alt="First Prompt" src="https://github.com/ingshtrom/hub-tool/blob/osx-install-screencast/media/osx_install_first_prompt.png" width="400" />
2. Open the "Security and Privacy" in System Preferences; choose the "General" tab. Allow the
   `hub-tool` security prompt at the bottom of the dialog
   <br />
   <img alt="Security and Privacy Dialog" src="https://github.com/ingshtrom/hub-tool/blob/osx-install-screencast/media/osx_install_security_and_privacy.png" width="400" />
3. Run the binary again `./hub-tool_darwin_amd64` and then accept the last
   prompt
   <br />
   <img alt="Last Prompt, Phew!" src="https://github.com/ingshtrom/hub-tool/blob/osx-install-screencast/media/osx_install_last_prompt.png" width="400" />

### Windows
TODO

### Linux
Download the binary from GitHub releases and then run `chmod +x ./hub-tool_linux_amd64`.


## Contributing

### Install from source
Just run `make install`, all the build is containerized. It will copy the hub tool binary to your
`/usr/local/bin` directory.
```shell script
$ make install
```

## Use it

```shell script
$ hub-tool ls nginx
TAG                    DIGEST                                                                    LAST UPDATE         SIZE
latest                 sha256:aff269ec296daeab62055236b6815322d6ae0752f6877e18b39261903463e0fc   3 days ago          58.95MB
stable-perl            sha256:8af9938e3a7afbabb6845864f305b84cdabd0f55c71ab6664d2b5385d77cb0fd   3 days ago          61.9MB
stable                 sha256:3f83cb7f711e08caed94b84edce5e4349b15e41158f027c37f547d03bb1f5dd2   3 days ago          50.13MB
perl                   sha256:fdea5e5cd991bf924cda48691cf693b646ce74b249d08c69e31054c224ffe422   3 days ago          60.41MB
mainline-perl          sha256:e89f3e62fc049f14beb5ea47897735ee426c0d85510d60e74446e866ec469386   3 days ago          70.12MB
mainline               sha256:794275d96b4ab96eeb954728a7bf11156570e8372ecd5ed0cbc7280313a27d19   3 days ago          53.5MB
1.19.2-perl            sha256:e89f3e62fc049f14beb5ea47897735ee426c0d85510d60e74446e866ec469386   3 days ago          70.12MB
1.19.2                 sha256:aff269ec296daeab62055236b6815322d6ae0752f6877e18b39261903463e0fc   3 days ago          58.95MB
1.19-perl              sha256:fdea5e5cd991bf924cda48691cf693b646ce74b249d08c69e31054c224ffe422   3 days ago          60.41MB
1.19                   sha256:aff269ec296daeab62055236b6815322d6ae0752f6877e18b39261903463e0fc   3 days ago          58.95MB
1.18.0-perl            sha256:59bdcdca6a76d3d295340f6189d65438cf5bdf767b893d1dba2fe76d0684e8b1   3 days ago          57.11MB
1.18.0                 sha256:48d22c8ecc16fa5da62dfb8de2d7c3f8ce9765df0678a4bc37556bac78a58ed0   3 days ago          53.42MB
1.18-perl              sha256:c6e5420e8a9ad4af82a2768f1cd8f7fd85da306eb92db24e9299207550a891d6   3 days ago          63.03MB
1.18                   sha256:91b74b601750353da4d76d059f21c0266799b28a6a3109b0d22c8fdbebc11c51   3 days ago          51.67MB
1-perl                 sha256:22fd43073a743547d21f3b845bdd55207425357dfb6b5ec6b33723cfcd8c7135   3 days ago          57.19MB
1                      sha256:c1d96b60af9efaf36f057e92c627fc721c9bd466a0ee19d3ff35c031c33bd0b7   3 days ago          51.72MB
mainline-alpine-perl   sha256:506fbc2cf89e768715c8f805dae8a71564a8045b01ba6d3b57fdbc53e357d037   5 weeks ago         16.62MB
mainline-alpine        sha256:3d4c3485cf8af9c0e38718409918ed6255caa32d6867cf667a7339b0c5a5641e   5 weeks ago         10.43MB
alpine-perl            sha256:b69f59203a518f1a1759ba8cc134fc144ebe87a834ee23da034790b793f7139b   5 weeks ago         18.65MB
alpine                 sha256:3d4c3485cf8af9c0e38718409918ed6255caa32d6867cf667a7339b0c5a5641e   5 weeks ago         10.43MB
1.19.2-alpine-perl     sha256:39c2f1440373d646484f34971c7ffe69744d4b4bc921e2f4d7b19cb0ffe71406   5 weeks ago         18.92MB
1.19.2-alpine          sha256:10bdb5cdba74478710feeb55a38b74fd57b5ea3f69072e1f8250af6c40b0fcb8   5 weeks ago         10.23MB
1.19-alpine-perl       sha256:4419334d3098762f669318b1a72271df9a6d7f34eec0fd08d0ca220d9b2ea88b   5 weeks ago         18.4MB
1.19-alpine            sha256:4635b632d2aaf8c37c8a1cf76a1f96d11b899f74caa2c6946ea56d0a5af02c0c   5 weeks ago         9.61MB
1-alpine-perl          sha256:9dcd7cd6567cd203afdd791ce770f5305b87b1ab588799ca9fbabce2c6cc44d3   5 weeks ago         19.13MB
```
