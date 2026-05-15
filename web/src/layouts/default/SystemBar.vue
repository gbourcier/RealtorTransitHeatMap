<template>
  <v-system-bar
    color="#262626"
    height="56"
  >
    <div class="d-flex align-center">
      <v-app-bar-nav-icon
        class="d-inline-block d-md-none"
        @click="emit('click:toggle')"
      />

      <div class="text-body-1 font-weight-black ms-4 text-white">
        RealtorTransitHeatMap
      </div>
    </div>

    <v-spacer />
  </v-system-bar>
</template>

<script lang="ts" setup>
const emit = defineEmits(['click:toggle'])
</script>
