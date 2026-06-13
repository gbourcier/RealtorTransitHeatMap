<template>
  <v-main :scrollable="scrollable">
    <router-view v-slot="{ Component }">
      <keep-alive include="Listings">
        <component :is="Component" />
      </keep-alive>
    </router-view>
  </v-main>
</template>

<script lang="ts" setup>
withDefaults(defineProps<{ scrollable?: boolean }>(), {
  scrollable: false,
})
</script>
