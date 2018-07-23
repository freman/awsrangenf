<template>
  <v-container fluid>

    <v-toolbar>
      <v-icon>settings</v-icon>
      <v-toolbar-title>Daemon Configuration</v-toolbar-title>
    </v-toolbar>

    <v-alert :value="readOnly" color="warning">
      Configuration is read only and cannot be edited
    </v-alert>
    <v-alert @input="err=''" :value="err != ''" color="error" dismissible transition="scale-transition">
      {{err}}
    </v-alert>
    <v-alert @input="success=''" :value="success != ''" color="success" dismissible transition="scale-transition">
      {{success}}
    </v-alert>

    <v-list two-line>
      <v-list-tile>
         <v-text-field
            :readonly="readOnly"
            hint="URL to retrieve the AWS IP Address Ranges data from"
            label="URL for ip-ranges.json"
            v-model="config.URL"
            :rules="[rules.required, rules.url]">
          </v-text-field>
      </v-list-tile><v-list-tile>
          <v-text-field
            :readonly="readOnly"
            hint="Timeout for all requests, supports suffixes of h, m, s. eg. 30s"
            label="Timeout"
            v-model="config.Timeout"
            :rules="[rules.required, rules.duration]">
          </v-text-field> 
      </v-list-tile><v-list-tile>
          <v-switch
            :readonly="readOnly"
            hint="Enable extra IPv6 routes"
            label="IPv6"
            v-model="config.IPv6">
          </v-switch>
      </v-list-tile>
      <v-subheader>IP Route</v-subheader>
      <v-divider />
      <v-list-tile>
        <v-text-field
          :readonly="readOnly"
          hint="Netfilter routing table 1-253"
          label="Routing Table"
          v-model="config.Route.Table"
          :rules="[rules.required, rules.table]">
        </v-text-field> 
      </v-list-tile><v-list-tile>
        <v-text-field
          :readonly="readOnly"
          hint="Gateway IP to store routes with, will be set automatically if 0.0.0.0 is specified"
          label="Default Gateway"
          v-model="config.Route.Gateway"
          :rules="[rules.required, rules.ip]">
        </v-text-field> 
      </v-list-tile>
      <v-subheader>Web Hook (SNS)</v-subheader>
      <v-divider />
      <v-list-tile>
          <v-switch
            :readonly="readOnly"
            hint="Enable webhook requests"
            label="Enabled"
            v-model="config.Webhook.Enabled">
          </v-switch>
      </v-list-tile><v-list-tile>
          <v-text-field
            :readonly="readOnly"
            hint="Secret key to include as part of the SNS callback url"
            label="Key/Token"
            v-model="config.Webhook.Key"
            append-icon="sync"
            @click:append="generateKey">
          </v-text-field>
      </v-list-tile>
      <v-subheader>Polling</v-subheader>
      <v-divider />
      <v-list-tile>
          <v-switch
            :readonly="readOnly"
            hint="Enable background polling"
            label="Enabled"
            v-model="config.Polling.Enabled">
          </v-switch>
      </v-list-tile><v-list-tile>
          <v-text-field
            :readonly="readOnly"
            hint="Polling interval, supports suffixes of h, m, s. eg. 30s"
            label="Interval"
            v-model="config.Polling.Interval"
            :rules="[rules.duration]">
          </v-text-field>
      </v-list-tile>
      <v-divider />
      <v-list-tile>
        <v-spacer />
        <v-btn color="primary" @click.stop="save()">Save</v-btn>
      </v-list-tile>
    </v-list>
  </v-container>
</template>

<script>
export default {
  name: "config",
  data() {
    return {
      readOnly: false,
      err: "",
      success: "",
      config: {
        Route: {},
        Webhook: {},
        Polling: {}
      },
      rules: {
        required: value => !!value || "Required.",
        url: value => {
          const pattern = /^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z0-9@:%_+.~#?&//=]*)$/;
          return pattern.test(value) || "Invalid URL.";
        },
        duration: value => {
          const pattern = /^(?:\d+h)(?:\d+m)?(?:\d+s)?|(?:\d+h)?(?:\d+m)(?:\d+s)?|(?:\d+h)?(?:\d+m)?(?:\d+s)|\d+$/;
          return pattern.test(value) || "Invalid Duration.";
        },
        table: value => {
          const pattern = /^\d+$/;
          let valint = parseInt(value, 10);
          return (
            (pattern.test(value) && valint > 0 && valint < 253) ||
            "Invalid table."
          );
        },
        ip: value => {
          const pattern = /^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$/;
          return pattern.test(value) || "Invalid IP.";
        }
      }
    };
  },
  methods: {
    generateKey() {
      const possible =
        "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_";

      this.config.Webhook.Key = "";
      for (var i = 0; i < 15; i++)
        this.config.Webhook.Key += possible.charAt(
          Math.floor(Math.random() * possible.length)
        );
    },
    save() {
      if (!this.readOnly) {
        this.axios
          .post("config", this.config)
          .then(response => {
            this.config = response.data.Config;
            this.success = "Configuration saved and reloaded";
          })
          .catch(error => {
            this[err] = error.response.data;
          });
      }
    }
  },
  beforeMount() {
    this.axios.get("config").then(response => {
      this.readOnly = response.data.ReadOnly;
      this.config = response.data.Config;
    });
  }
};
</script>

