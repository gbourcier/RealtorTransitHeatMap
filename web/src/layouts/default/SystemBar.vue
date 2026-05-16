<template>
  <v-system-bar
    color="#262626"
    height="56"
  >
    <div class="d-flex align-center system-bar__lockup">
      <v-app-bar-nav-icon
        class="d-inline-block d-lg-none"
        @click="emit('click:toggle')"
      />

      <div class="system-bar__brand text-body-1 font-weight-black ms-4 text-white d-none d-lg-block">
        RealtorTransitHeatMap
      </div>

      <div class="system-bar__title text-body-1 font-weight-medium ms-2 text-white d-lg-none">
        {{ pageTitle }}
      </div>
    </div>

    <v-spacer />
  </v-system-bar>
</template>

<script lang="ts" setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const emit = defineEmits(['click:toggle'])

const route = useRoute()
const pageTitle = computed(() => (route.meta?.title as string) ?? 'RealtorTransitHeatMap')
</script>

<style scoped>
.system-bar__lockup {
  min-width: 0;
}

.system-bar__title {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
