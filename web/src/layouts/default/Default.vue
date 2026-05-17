<template>
  <v-app>
    <default-system-bar @click:settings="settingsDrawer = !settingsDrawer" />

    <default-drawer v-model="settingsDrawer" />

    <default-view />
  </v-app>
</template>

<script lang="ts" setup>
import DefaultDrawer from './Drawer.vue'
import DefaultSystemBar from './SystemBar.vue'
import DefaultView from './View.vue'

import { shallowRef } from 'vue'

const settingsDrawer = shallowRef<boolean | null>(false)
</script>
