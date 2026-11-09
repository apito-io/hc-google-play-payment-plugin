# Google Play payment plugin

Google Play Billing and Real-time developer notifications for a project's tenant plans.

Credentials stay in the plugin's project env. Install from the Apito plugin catalog, then activate it for the project and save the package name and service account on the payment settings page.

The app still calls `POST /system/v1/project-payment/projects/:projectId/providers/google_play/verify`. The engine dispatches that call to this plugin.
