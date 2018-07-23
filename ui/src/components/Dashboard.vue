<template>
  <v-container fluid>
    <v-toolbar>
      <v-icon>dashboard</v-icon>
      <v-toolbar-title>Dashboard</v-toolbar-title>
    </v-toolbar>
    <v-alert :value="!Bootstrap.Finished">
      Startup failure while "{{Bootstrap.Label}}" got "{{Bootstrap.Error}}"
    </v-alert>
    <v-card>
      <v-container fluid grid-list-lg>
        <v-layout row wrap>
          <v-flex v-for="(data, name) in Cards" xs4 :key="name">
            <v-card>
              <v-card-title primary-title>
                {{name}}
              </v-card-title>
              <v-card-text>
                <div class="headline">{{data}}</div>
              </v-card-text>
            </v-card>
          </v-flex>
        </v-layout>
      </v-container>
    </v-card>
    <v-card>
      <v-list dense>
        <v-list-tile v-for="log in Logs" :key="log">
          <v-list-tile-content>
            {{log}}
          </v-list-tile-content>
        </v-list-tile>
      </v-list>
    </v-card>
  </v-container>
</template>

<script>
export default {
  name: "dashboard",
  data() {
    return {
      Bootstrap: {},
      Cards: {},
      Logs: []
    };
  },
  beforeMount() {
    this.axios.get("dashboard").then(response => {
      this.Bootstrap = response.data.Bootstrap;
      this.Cards = response.data.Cards;
      this.Logs = response.data.Logs.reverse();
    });
  }
};
</script>

