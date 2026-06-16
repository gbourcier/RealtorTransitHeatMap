<template>
  <v-app>
    <template v-if="!isPublicRoute">
      <default-system-bar
        :settings-active="isSettingsActive"
        :show-back="showBack"
        :show-panel-toggle="route.name === 'listings'"
        @click:brand="settingsDrawer = false"
        @click:settings="onToggleSettings"
        @click:back="onBackToMap"
        @click:panel-toggle="onToggleListingsPanel"
      />

      <default-drawer
        v-if="!mobile"
        :model-value="isSettingsRoute || settingsDrawer"
        :permanent="isSettingsRoute"
        @update:model-value="(v) => (settingsDrawer = v)"
      />
    </template>

    <default-view :scrollable="isSettingsRoute" />
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
const isPublicRoute = computed(() => route.meta?.public === true)
const isSettingsRoute = computed(() => route.meta?.settings === true)
const settingsDrawer = shallowRef(false)
const isSettingsActive = computed(() => isSettingsRoute.value || (!mobile.value && settingsDrawer.value))
const showBack = computed(() => isSettingsRoute.value || route.name === 'favorites')

watch(isSettingsRoute, (active) => {
  if (active) settingsDrawer.value = false
})

watch(mobile, (active) => {
  if (active) settingsDrawer.value = false
})

function onToggleSettings() {
  if (mobile.value || isSettingsRoute.value) return
  settingsDrawer.value = !settingsDrawer.value
}

function onBackToMap() {
  settingsDrawer.value = false
  router.push('/listings')
}

function onToggleListingsPanel() {
  window.dispatchEvent(new CustomEvent('listings:toggle-panel'))
}
</script>
