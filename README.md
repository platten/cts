#CTS Server Setup Documentation

##Workstation initial setup

###Install core dev apps
* Install VirtualBox: https://www.virtualbox.org/wiki/Downloads
* Install VirtualBox extension
* Install iTerm2: https://www.iterm2.com/downloads.html
* Install Atom: https://atom.io/
* Install Mac github client: https://mac.github.com/
* Install Xcode from Apple App Store
* Open iTerm2
* Create development directory
`mkdir $HOME/Development`
* Install Xcode CLI tools from within Xcode
`xcode-select --install`
* Install Google Cloud SDK:
`curl https://sdk.cloud.google.com | bash`
* Restart shell
* Install preview for Google Cloud SDK
`gcloud components update preview`
* Set version of Google Cloud SDK
`gcloud config set component_manager/fixed_sdk_version 0.9.64`
* Update component versions
`cloud components update`
* Install Homebrew:
`ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"`
* Install tools from homebrew
`brew install git homebrew/dupes/openssh autoconf automake libtool pkg-config go boot2docker`
* Optionally install:
`brew install mc tmux iftop htop mosh watch mercurial wget ffmpeg youtube-dl multitail`


###Configure Go
* Open iTerm2
* Create directory for for $GOPATH:
`mkdir -p $HOME/Development/go`
* Add development root to environment and refresh shell
  * For bash:
  `cat >> ~/.bashrc <<DELIM`
  `export DEVROOT=$HOME/Development`
  `export GOPATH=$DEVROOT/go`
  `export APPENGINE_SDK=$DEVROOT/google-cloud-sdk/platform/google_appengine`
  `export PATH=$PATH:$APPENGINE_SDK:$GOPATH/bin`
  `source "$HOME/Development/google-cloud-sdk/path.bash.inc"`
  `source "$HOME/Development/google-cloud-sdk/completion.bash.inc"`
  `export DOCKER_HOST=tcp://192.168.59.103:2375`
  `unset DOCKER_TLS_VERIFY`
  `unset DOCKER_CERT_PATH`
  `DELIM`
  `source ~/.bashrc`
  * For zsh:
  `cat >> ~/.zshrc <<DELIM`
  `export DEVROOT=$HOME/Development`
  `export GOPATH=$DEVROOT/go`
  `export APPENGINE_SDK=$DEVROOT/google-cloud-sdk/platform/google_appengine`
  `export PATH=$PATH:$APPENGINE_SDK:$GOPATH/bin`
  `source "$HOME/Development/google-cloud-sdk/path.zsh.inc"`
  `source "$HOME/Development/google-cloud-sdk/completion.zsh.in"`
  `if [[ -s "${ZDOTDIR:-$HOME}/.zprezto/init.zsh" ]]; then`
  `  source "${ZDOTDIR:-$HOME}/.zprezto/init.zsh"`
  `fi`
  `export DOCKER_HOST=tcp://192.168.59.103:2375`
  `unset DOCKER_TLS_VERIFY`
  `unset DOCKER_CERT_PATH`
  `DELIM`
  `source ~/.zshrc`
   `echo 'fi' >> ~/.zshrc`
  `source ~/.zshrc`
* Install godep
`go get github.com/tools/godep`
* Link Google Cloud SDK with $GOPATH
`ln -s $APPENGINE_SDK/goroot/src/appengine_internal $GOPATH/src/`
`ln -s $APPENGINE_SDK/goroot/pkg/darwin_amd64_appengine/appengine_internal $GOPATH/pkg/darwin_amd64/`

###Install Protocol Buffers
* Make build directory & cd into it
`mkdir $DEVROOT/build; cd $DEVROOT/build`
* Download protobuf code
`git clone https://github.com/google/protobuf.git`
* Build & install the protobuf library
`cd protobuf`
`./autogen.sh`
`./configure`
`make`
`make check`
`make install`
* Install protobuf generator for Go
`go get -a github.com/golang/protobuf/protoc-gen-go`
`go get -a github.com/golang/protobuf/proto`


###Configure Managed VM environment
* Initialize docker
`boot2docker init`
* Start docker
`boot2docker up`
* SSH to boot2docker
`boot2docker ssh`
* Disable TLS in boot2docker host
`sudo -i`
`echo 'DOCKER_TLS=no' > /var/lib/boot2docker/profile`
`echo 'EXTRA_ARGS="--insecure-registry 10.2.4.201"' >> /var/lib/boot2docker/profile`
* Restart docker service in boot2docker
`/etc/init.d/docker restart`
* Exit from sudo shell
`exit`
* Exit from boot2docker vm
`exit`
* Verify docker is running
```boot2docker status
docker ps```
* Optional: Download docker images
`docker pull gcr.io/google_appengine/python-compat`
`docker pull gcr.io/google_appengine/go-compat`
* Optional: Verify docker images are downloaded
`docker images`
* Login to google cloud
`gcloud auth login`
* Set google cloud project
`getcloud config set project <project name>`

###Git Setup
* Set up Git:
https://help.github.com/articles/set-up-git/
* Set up credentials for HTTPS access to GitHub
https://help.github.com/articles/caching-your-github-password-in-git/


###CTS initial checkout
* Enter development directory
`cd $DEVROOT`
*  Check out CTS repository from GitHub:
`git clone https://github.com/theorangechefco/cts.git`
* Enter the CTS repository
`cd cts`
* Initialize and update submodules
`git submodule init`
`git submodule update`
*  Add cts bin directory to PATH
`echo 'PATH=$PATH:$DEVROOT/cts/bin'`
* Reload environment
  * For bash:
  `source ~/.bashrc`
  * For zsh:
  `source ~/.zshrc`

###Locally run sample gRPC server
* Enter appropriate project directory
`cd $DEVROOT/cts/grpc-go-template`
* Start local dev environment
`go_script.sh --run`

> ####Note:
> The template code is based on the gRPC greeter server code available here:
> https://github.com/grpc/grpc-common/tree/master/go

###Access locally running sample gRPC server
* Open new iTerm window
* Install grpc-common
`go get github.com/grpc/grpc-common`
* Enter greeter_client directory
`cd $GOPATH/src/github.com/grpc/grpc-common/go/greeter_client`
* Run local client
`go run main.go`
