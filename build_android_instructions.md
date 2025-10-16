# Building Android APK for DSB - Delta Sharing Browser

This guide provides step-by-step instructions for building an Android APK of the DSB application.

## Prerequisites

Before building for Android, you need to set up your development environment with the following tools:

### 1. Install Go

Ensure you have Go 1.22.7 or later installed:
```bash
go version
```

### 2. Install Android Studio

Download and install Android Studio from: https://developer.android.com/studio

### 3. Install Android SDK

1. Open Android Studio
2. Go to **Settings/Preferences** → **Appearance & Behavior** → **System Settings** → **Android SDK**
3. Install the following:
   - Android SDK Platform (API 21 or higher)
   - Android SDK Build-Tools
   - Android SDK Platform-Tools

### 4. Install Android NDK

1. In Android Studio, go to **Settings/Preferences** → **Appearance & Behavior** → **System Settings** → **Android SDK**
2. Click on the **SDK Tools** tab
3. Check **NDK (Side by side)**
4. Click **Apply** to install

### 5. Set Environment Variables

Add these to your shell configuration file (`~/.bashrc`, `~/.zshrc`, or similar):

```bash
# Android SDK
export ANDROID_HOME=$HOME/Library/Android/sdk

# Android NDK (replace <version> with your installed version)
export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/<version>

# Add to PATH
export PATH=$PATH:$ANDROID_HOME/platform-tools
export PATH=$PATH:$ANDROID_HOME/tools
```

To find your NDK version:
```bash
ls $HOME/Library/Android/sdk/ndk/
```

Apply the changes:
```bash
source ~/.zshrc  # or ~/.bashrc
```

Verify the setup:
```bash
echo $ANDROID_HOME
echo $ANDROID_NDK_HOME
```

### 6. Install Fyne CLI Tool

The Makefile will automatically install this, but you can also install it manually:
```bash
go install fyne.io/fyne/v2/cmd/fyne@latest
```

Ensure `$GOPATH/bin` is in your PATH:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

## Building the Android APK

### Using the Makefile

Once prerequisites are installed, building is simple:

```bash
# Build Android APK
make android
```

This will create a file named `dsb.apk` in your project directory.

### For Google Play Store (AAB format)

```bash
# Build Android AAB (Android App Bundle)
make android-aab
```

This will create a file named `dsb.aab` in your project directory.

### Manual Build (without Makefile)

If you prefer to build manually:

```bash
# Install fyne CLI
go install fyne.io/fyne/v2/cmd/fyne@latest

# Build APK
fyne package -os android -appID com.example.dsb -icon Icon.png -name dsb

# Or build AAB for Google Play
fyne package -os android/aab -appID com.example.dsb -icon Icon.png -name dsb
```

## Configuration Options

You can customize the build by editing the Makefile variables:

- **APP_NAME**: Application name (default: `dsb`)
- **APP_ID**: Android package ID (default: `com.example.dsb`)
- **VERSION**: Application version (default: `1.0.0`)
- **ICON**: Icon file path (default: `Icon.png`)

Example customization in Makefile:
```makefile
APP_NAME = MyDSB
APP_ID = com.mycompany.dsb
VERSION = 1.0.0
ICON = assets/icon.png
```

## Installing on Android Device

### Via USB (ADB)

1. Enable Developer Options on your Android device
2. Enable USB Debugging
3. Connect your device via USB
4. Install the APK:

```bash
adb install dsb.apk
```

### Via File Transfer

1. Copy `dsb.apk` to your Android device
2. Open the file on your device
3. Allow installation from unknown sources if prompted
4. Tap "Install"

## Signing the APK for Release

For distribution, you should sign your APK:

### 1. Generate a Keystore

```bash
keytool -genkey -v -keystore my-release-key.jks -keyalg RSA -keysize 2048 -validity 10000 -alias my-key-alias
```

### 2. Sign the APK

```bash
jarsigner -verbose -sigalg SHA256withRSA -digestalg SHA-256 -keystore my-release-key.jks dsb.apk my-key-alias
```

### 3. Optimize with zipalign

```bash
zipalign -v 4 dsb.apk dsb-aligned.apk
```

## Troubleshooting

### Common Issues

#### "ANDROID_HOME not set"
```bash
export ANDROID_HOME=$HOME/Library/Android/sdk
```

#### "NDK not found"
```bash
# Check installed NDK versions
ls $HOME/Library/Android/sdk/ndk/

# Set NDK_HOME with correct version
export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/26.1.10909125
```

#### "fyne: command not found"
```bash
go install fyne.io/fyne/v2/cmd/fyne@latest
export PATH=$PATH:$(go env GOPATH)/bin
```

#### Build fails with "package not found"
```bash
make deps  # or go mod download && go mod tidy
```

## Additional Resources

- **Fyne Documentation**: https://docs.fyne.io/
- **Fyne Mobile Packaging**: https://docs.fyne.io/started/mobile
- **Android Developer Guide**: https://developer.android.com/guide
- **Go Mobile**: https://github.com/golang/mobile

## Build All Platforms

To build for all supported platforms:
```bash
make build-all
```

This will create builds for:
- Linux
- Windows
- macOS
- iOS
- Android

## Cleaning Up

Remove all build artifacts:
```bash
make clean
```

## Support

For issues specific to:
- **DSB Application**: Create an issue in the project repository
- **Fyne Framework**: Visit https://github.com/fyne-io/fyne
- **Android Development**: Visit https://developer.android.com/
