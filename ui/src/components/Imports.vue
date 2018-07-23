<template>
  <v-container fluid>

    <v-toolbar>
      <v-icon>cloud_download</v-icon>
      <v-toolbar-title>Importing from AWS</v-toolbar-title>
    </v-toolbar>
    <v-alert @input="removeError=''" dismissible type="error" transition="slide-y-transition" :value="removeError!==''">{{removeError}}</v-alert>

    <v-card>
      <v-card-text>
        <div class="text-xs-center">
          <v-chip v-for="filter in imports.Filter" :key="filter" close @input="remove(filter)">{{filter}}</v-chip>
        </div>
      </v-card-text>
      <v-card-actions>
        <div>Importing {{imports.Count}} out of {{imports.Total}} prefixes</div>
        <v-spacer></v-spacer>
        <v-speed-dial v-model="fab" bottom right open-on-hover transition="scale-transition" direction="left">
          <v-btn slot="activator" v-model="fab" color="green" fab><v-icon>add_circle</v-icon><v-icon>add_circle_outline</v-icon></v-btn>
          <v-btn large @click.stop="selectByRegion()">By Region</v-btn>
          <v-btn large @click.stop="selectByService()">By Service</v-btn>
        </v-speed-dial>
      </v-card-actions>
    </v-card>

    <v-bottom-sheet v-model="sheet" inset persistent>
      <v-stepper v-model="e1">
        <v-stepper-header>
          <v-stepper-step editable :complete="e1 > 1" step="1">Select {{pick1}}</v-stepper-step>
          <v-divider></v-divider>
          <v-stepper-step :complete="e1 > 2" step="2">Select {{pick2}}</v-stepper-step>
        </v-stepper-header>
        <v-alert @input="addError=''" dismissible type="error" transition="slide-y-transition" :value="addError!==''">{{addError}}</v-alert>
        <v-stepper-items>
          <v-stepper-content step="1">
            <v-card class="mb-5">
              <v-combobox :items="step1" dense autofocus v-model="chosen1" :label="pick1">
              </v-combobox>
            </v-card>
            <v-btn color="primary" @click.stop="e1 = chosen1 === '' ? 1 : 2">Continue</v-btn>
            <v-btn flat @click.stop="cancel()"> Cancel </v-btn>
          </v-stepper-content>
          <v-stepper-content step="2">
            <v-card class="mb-5">
              <v-combobox :items="step2" dense autofocus v-model="chosen2" :label="pick2">
              </v-combobox>
            </v-card>
            <v-btn color="primary" @click.stop="doneSelect()">Done</v-btn>
            <v-btn flat @click.stop="cancel()"> Cancel </v-btn>
          </v-stepper-content>
        </v-stepper-items>
      </v-stepper>
    </v-bottom-sheet>
  </v-container>
</template>

<script>
export default {
  name: "imports",
  data() {
    return {
      sheet: false,
      fab: false,
      imports: {
        Count: 0,
        Total: 0,
        Filter: [],
        RegionToService: {},
        ServiceToRegion: {}
      },

      addError: "",
      removeError: "",
      e1: 1,
      chosen1: "",
      chosen2: "",
      pick1: "",
      pick2: "",
      pickFrom: {},
      doneFunc() {}
    };
  },
  computed: {
    step1() {
      return Object.keys(this.pickFrom);
    },
    step2() {
      return ["*"].concat(this.pickFrom[this.chosen1]);
    }
  },
  methods: {
    resetSelect(first, second) {
      this.e1 = 1;
      this.chosen1 = "";
      this.chosen2 = "";
      this.pick1 = first;
      this.pick2 = second;
      this.sheet = true;
    },
    selectByRegion() {
      this.pickFrom = this.imports.RegionToService;
      this.doneFunc = () => {
        return this.chosen1 + ":" + this.chosen2;
      };
      this.resetSelect("Region", "Service");
    },
    selectByService() {
      this.pickFrom = this.imports.ServiceToRegion;
      this.doneFunc = () => {
        return this.chosen2 + ":" + this.chosen1;
      };
      this.resetSelect("Service", "Region");
    },
    doneSelect() {
      if (this.chosen1 === "" || this.chosen2 === "") {
        return;
      }

      let tmp = this.imports.Filter ? this.imports.Filter.slice() : [];
      let value = this.doneFunc();

      if (tmp.indexOf(value) > -1) {
        return;
      }

      tmp.push(value);
      this.submit(tmp, "addError", () => {
        this.sheet = false;
      });
    },
    remove(filter) {
      let tmp = this.imports.Filter ? this.imports.Filter.slice() : [];
      tmp.splice(tmp.indexOf(filter), 1);
      this.submit(tmp, "removeError");
    },
    cancel() {
      this.addError = "";
      this.sheet = false;
      this.e1 = 1;
    },
    submit(filter, err, andthen) {
      this.axios
        .post("imports", filter)
        .then(response => {
          this.imports.Filter = response.data.Filter;
          this.imports.Total = response.data.Total;
          this.imports.Count = response.data.Count;
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
    this.axios.get("imports").then(response => {
      this.imports = response.data;
    });
  }
};
</script>

