<template>
  <v-app>
    <default-system-bar
      :settings-active="isSettingsActive"
      @click:settings="onToggleSettings"
    />

    <default-drawer
      :model-value="(isSettingsRoute && !mobile) || settingsDrawer"
      :permanent="isSettingsRoute && !mobile"
      @update:model-value="(v) => (settingsDrawer = v)"
    />

    <default-view />
  </v-app>
</template>

<script lang="ts" setup>
import DefaultDrawer from './Drawer.vue'
import DefaultSystemBar from './SystemBar.vue'
import DefaultView from './View.vue'

import { computed, shallowRef, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useDisplay } from 'vuetify'

const route = useRoute()
const router = useRouter()
const { mobile } = useDisplay()
const isSettingsRoute = computed(() => route.meta?.settings === true)
const settingsDrawer = shallowRef<boolean | null>(false)
const isSettingsActive = computed(() => isSettingsRoute.value || !!settingsDrawer.value)

watch(isSettingsRoute, (active) => {
  if (active) settingsDrawer.value = false
})

function onToggleSettings() {
  if (isSettingsRoute.value) {
    settingsDrawer.value = false
    router.push('/listings')
    return
  }
  settingsDrawer.value = !settingsDrawer.value
}
</script>
