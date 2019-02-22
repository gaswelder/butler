#!/bin/sh

APP_ARCHIVE_PATH="myapp.xcarchive"
IPA_PATH="myappIPA"

export PATH="${PATH}:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/Library/Internet Plug-Ins/JavaAppletPlugin.plugin/Contents/Home/bin"
export LC_ALL=en_US.UTF-8
export LANG=en_US.UTF-8
export LANGUAGE=en_US.UTF-8

rm -fr ~/Library/MobileDevice/Provisioning\ Profiles/
rm -fr "${IPA_PATH}"
rm -fr "${APP_ARCHIVE_PATH}"

/usr/libexec/PlistBuddy -c "Set :CFBundleShortVersionString ${TAG_VERSION}" myapp/Info.plist
/usr/libexec/PlistBuddy -c "Set :CFBundleVersion ${BUNDLE_VERSION}" myapp/Info.plist
/usr/libexec/PlistBuddy -c "Set :CFBundleIdentifier ${PROJECT_ID}" myapp/Info.plist
/usr/libexec/PlistBuddy -c "Set :CFBundleDisplayName ${PROJECT_NAME}" myapp/Info.plist
/usr/libexec/PlistBuddy -c "Set :MixpanelAPIToken ${MixpanelAPIToken}" myapp/Info.plist

security -v unlock-keychain -p "${BUILD_CERT_PASSWD}" "/Users/macslave/Library/Keychains/login.keychain-db"
security -v unlock-keychain -p "${BUILD_CERT_PASSWD}" "/Users/macslavenew/Library/Keychains/login.keychain-db"
fastlane build_adhoc
fastlane build_store
rm *.mobileprovision

cp -fr "manifest_template.plist" "${IPA_PATH}/myapp.plist"
/usr/libexec/PlistBuddy -c "Set :items:0:metadata:bundle-identifier ${PROJECT_ID}" ${IPA_PATH}/myapp.plist
/usr/libexec/PlistBuddy -c "Set :items:0:metadata:bundle-version ${TAG_VERSION}" ${IPA_PATH}/myapp.plist
/usr/libexec/PlistBuddy -c "Set :items:0:metadata:title ${PROJECT_NAME}" ${IPA_PATH}/myapp.plist
/usr/libexec/PlistBuddy -c "Set :items:0:assets:0:url chAngEmE001" ${IPA_PATH}/myapp.plist


${IPA_PATH}/myapp.plist

"mkdir -p ${FINAL_PATH}; cd ${TAG_PATH}; 
put ${SCRIPT_DIR}/.tag; cd ${FINAL_PATH}; put ${IPA_PATH}/myapp.plist; put ${IPA_PATH}/myapp.ipa; bye"
	
rm -fr ${SCRIPT_DIR}/.tag > /dev/null 2>&1
