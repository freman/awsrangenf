<template>
  <v-container fluid>

    <v-toolbar>
      <v-icon>layers</v-icon>
      <v-toolbar-title>Custom routes</v-toolbar-title>
    </v-toolbar>
    <v-alert @input="removeError=''" dismissible type="error" transition="slide-y-transition" :value="removeError!==''">{{removeError}}</v-alert>
    <v-card>
      <v-card-text>
        You can add custom routes here and they will be propogated to dependant networks, this is most useful for temporarily adding a route to test via AWS.
      </v-card-text>
    </v-card>
    <v-card>
      <v-card-text>
        <div class="text-xs-center">
          <v-chip v-for="net in custom" :key="net" close @input="remove(net)">{{net}}</v-chip>
        </div>
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn bottom right fab color="green" @click.stop="add()"><v-icon>add_circle</v-icon><v-icon>add_circle_outline</v-icon></v-btn>
      </v-card-actions>
    </v-card>

    <v-bottom-sheet v-model="sheet" inset persistent>
      <v-card>
        <v-alert @input="addError=''" dismissible type="error" transition="slide-y-transition" :value="addError!==''">{{addError}}</v-alert>
        <v-card-text>
          <p>Caution: It is possible to break all the things by adding a bad custom route, please use this carefully</p>
          <v-text-field dense autofocus v-model="addValue" label="Network CIDR (0.0.0.0/0)" :rules="[rules.required, rules.cidr]">></v-text-field>
        </v-card-text>
        <v-card-actions>
          <v-btn color="primary" @click.stop="done()">Done</v-btn>
          <v-btn flat @click.stop="cancel()"> Cancel </v-btn>
        </v-card-actions>
      </v-card>
    </v-bottom-sheet>
  </v-container>
</template>

<script>
export default {
  name: "custom",
  data() {
    return {
      sheet: false,
      fab: false,
      custom: [],

      addValue: "",
      addError: "",
      removeError: "",
      rules: {
        required: value => !!value || "Required.",
        cidr: value => {
          const pattern = /^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\/(\d+)$/;
          return (
            (pattern.test(value) && value !== "0.0.0.0/0") || "Invalid CIDR."
          );
        }
      }
    };
  },
  methods: {
    add() {
      this.addValue = "";
      this.sheet = true;
    },
    remove(net) {
      let tmp = this.custom ? this.custom.slice() : [];
      tmp.splice(tmp.indexOf(net), 1);
      this.submit(tmp, "removeError");
    },
    done() {
      let tmp = this.custom ? this.custom.slice() : [];
      if (tmp.indexOf(this.addValue) > -1) {
        return;
      }
      tmp.push(this.addValue);
      this.submit(tmp, "addError", () => {
        this.sheet = false;
        this.addValue = "";
      });
    },
    cancel() {
      this.addError = "";
    },
    submit(custom, err, andthen) {
      this.axios
        .post("custom", custom)
        .then(response => {
          this.custom = response.data ? response.data : [];
          if (andthen) {
            andthen();
          }
        })
        .catch(error => {
          this[err] = error.response.data;
        });
    }
  },
  beforeMount() {
    this.axios.get("custom").then(response => {
      this.custom = response.data ? response.data : [];
    });
  }
};
</script>

