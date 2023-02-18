# Github action to upload and publish extension to Chrome WebStore

Github action to automate Chrome extension upload and publishing to Chrome Web Store.

## Authentication to Chrome Web Store
You will need to create OAuth client with "https://www.googleapis.com/auth/chromewebstore" scope and store your refresh token as Github repository secret.

## Usage
Add build task to your workflow. 
Supported actions: 
- "upload" - upload a new extension version
- "uploadPublish" - upload and publish extension at the same time
```yml
- name: Upload to webstore
  uses: ExtensionNinja/extension-publish@main
  with:
   action: upload
   extensionID: jcobahfcgpekhnfcplgojikapkkbgmkh
   clientID: ${{ secrets.GOOGLE_CLIENT_ID }}
   clientSecret: ${{ secrets.GOOGLE_CLIENT_SECRET }}
   clientRefreshToken: ${{ secrets.GOOGLE_REFRESH_TOKEN }}
   extensionFile: ./extension.zip
```
