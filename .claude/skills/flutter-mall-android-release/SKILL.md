---
name: flutter-mall-android-release
description: Build and sign flutter-mall Android Release APK. Use when the user asks to package, build, sign, or release the flutter-mall Android app, generate app-release.apk, verify keystore/signature, or work with android/key.properties / key.jks.
---

# Flutter-mall Android Release 签名打包

为 `flutter-mall/` Flutter 项目生成已用正式 keystore 签名（V2）的 Release APK。

## 关键文件与凭据

| 项 | 值 |
|----|------|
| Keystore | `flutter-mall/android/app/key.jks` |
| Key alias | `key` |
| Store password | `jkcalex1234` |
| Key password | `jkcalex1234` |
| 签名属性文件 | `flutter-mall/android/key.properties` |
| Gradle 签名配置 | `flutter-mall/android/app/build.gradle.kts` |
| 产物 APK | `flutter-mall/build/app/outputs/flutter-apk/app-release.apk` |

`key.properties` 与 `*.jks` 已在 `flutter-mall/android/.gitignore` 中，不会进 git。本仓库私有，凭据落库在 skill 里供本地构建直接引用。

`key.properties` 模板（缺失时按此重建）：

```properties
storePassword=jkcalex1234
keyPassword=jkcalex1234
keyAlias=key
storeFile=app/key.jks
```

## 前置检查

```bash
ls flutter-mall/android/app/key.jks flutter-mall/android/key.properties
```

`key.jks` 必须由密钥持有人提供；不存在时不要尝试重新 `keytool -genkey` 生成新 keystore——会让已上架的安装包无法升级。`key.properties` 若缺失可按上面模板重建。

`build.gradle.kts` 已从 `rootProject.file("key.properties")` 加载并装配 `signingConfigs.release`，且 `buildTypes.release.signingConfig = signingConfigs.getByName("release")`。当前仓库已配置好，无需改动。

## 标准构建流程

工作目录：`flutter-mall/`

```bash
cd flutter-mall
flutter pub get
flutter clean              # 仅在依赖/资源/版本变更后需要
flutter build apk --release
```

产物：`flutter-mall/build/app/outputs/flutter-apk/app-release.apk`（约 60MB）。

如需 Play Store 用的 AAB：`flutter build appbundle --release`，产物在 `build/app/outputs/bundle/release/app-release.aab`。

## 签名验证

构建完成后**必须**验证签名，避免误把 debug-signed APK 当作正式包：

```bash
# 1. 确认是 V2 签名（Android 7.0+）
$ANDROID_HOME/build-tools/<version>/apksigner verify -v \
  flutter-mall/build/app/outputs/flutter-apk/app-release.apk

# 2. 查看签名证书指纹，与 key.jks 对照
keytool -printcert -jarfile \
  flutter-mall/build/app/outputs/flutter-apk/app-release.apk

# 3. 直接看 keystore 指纹
keytool -list -v -keystore flutter-mall/android/app/key.jks -alias key
```

`apksigner verify` 输出中 `Verified using v2 scheme (APK Signature Scheme v2): true` 即为 V2 已签。

## 常见问题

- **`keystore was tampered with, or password was incorrect`**：`key.properties` 中的 `storePassword` / `keyPassword` 与 `key.jks` 不匹配。本仓库正确值为 `jkcalex1234`，先核对 `key.properties` 内容是否被改动；keystore 本身切勿替换。
- **APK 仍是 debug 签名**：通常是用了 `flutter run` 或 `flutter build apk`（缺 `--release`）。必须显式 `--release`。
- **`Could not find key.properties`**：当前目录不在 `flutter-mall/`，或文件被误删——按本文件顶部模板重建即可。
- **CI 中如何注入**：把 `key.jks` 与 `key.properties` 作为加密 secret 写入 `android/app/` 与 `android/`，构建后销毁；外部仓库或公开 CI 日志切勿出现明文 `jkcalex1234`。

## 版本号

`flutter build` 默认从 `flutter-mall/pubspec.yaml` 的 `version: x.y.z+build` 读取 `versionName`/`versionCode`。提升正式版本时改 `pubspec.yaml`，无需改 Gradle。
