import { PLUGIN_ID, sdk } from "./host";
import ProviderTab from "./provider";

sdk().register({
  name: PLUGIN_ID,
  version: "0.1.0",
  displayName: "Google Play",
  routes: [],
  components: {
    slots: {
      "payment-providers": ProviderTab,
    },
  },
});
