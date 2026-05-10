<script setup lang="ts">
import { ref } from 'vue'

const drawer = ref(true)
</script>

<template>
  <v-app>
    <v-app-bar color="surface" flat border>
      <v-app-bar-nav-icon @click="drawer = !drawer" />
      <v-app-bar-title class="font-weight-bold">RealtorTransitHeatMap</v-app-bar-title>
    </v-app-bar>

    <v-navigation-drawer v-model="drawer" rail-width="64">
      <v-list nav>
        <v-list-subheader>DASHBOARDS</v-list-subheader>
        <v-list-item
          to="/"
          prepend-icon="mdi-view-dashboard-outline"
          title="Dashboard"
          value="dashboard"
        />
      </v-list>
    </v-navigation-drawer>

    <v-main class="bg-grey-lighten-4">
      <router-view />
    </v-main>
  </v-app>
</template>
