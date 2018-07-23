<template>
  <v-app id="inspire">
    <v-navigation-drawer v-model="drawer" app clipped fixed>
      <v-list>
        <template v-for="(route, index) in routes">
          <v-list-tile :to="route.to" :exact="route.exact" :key="index">
            <v-list-tile-action>
              <v-icon v-if="route.icon">{{route.icon}}</v-icon>
            </v-list-tile-action>
            <v-list-tile-content>
              {{route.name}}
            </v-list-tile-content>
          </v-list-tile>
        </template>
      </v-list>
    </v-navigation-drawer>
    <v-toolbar app clipped-left fixed>
      <v-toolbar-side-icon @click.stop="drawer = !drawer"></v-toolbar-side-icon>
      <v-toolbar-title>AWS Routes</v-toolbar-title>
    </v-toolbar>
    <v-content>
      <router-view/>
    </v-content>
    <v-footer></v-footer>
  </v-app>
</template>

<script>
export default {
  name: "App",
  data() {
    return {
      drawer: true,
      routes: [
        {
          exact: true,
          name: "Dashboard",
          to: { name: "dashboard" },
          icon: "dashboard"
        },
        {
          name: "AWS Imports",
          to: { name: "imports" },
          icon: "cloud_download"
        },
        {
          name: "Custom Routes",
          to: { name: "custom" },
          icon: "layers"
        },
        {
          name: "Daemon Configuration",
          to: { name: "daemon" },
          icon: "settings"
        }
      ]
    };
  }
};
</script>
