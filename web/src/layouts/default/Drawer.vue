<template>
  <v-navigation-drawer
    :model-value="modelValue"
    location="left"
    :temporary="!permanent"
    :permanent="permanent"
    width="260"
    @update:model-value="(v) => emit('update:modelValue', v)"
  >
    <div class="settings-drawer__header px-4 pt-4 pb-2">
      <div class="text-overline text-medium-emphasis">Settings</div>
    </div>

    <v-list nav density="comfortable">
      <v-list-item
        v-for="item in items"
        :key="item.to"
        :to="item.to"
        :title="item.title"
        :prepend-icon="item.icon"
        :subtitle="item.subtitle"
        color="secondary"
        rounded="lg"
      />
    </v-list>
  </v-navigation-drawer>
</template>

<script lang="ts" setup>
withDefaults(defineProps<{ modelValue: boolean | null; permanent?: boolean }>(), {
  permanent: false,
})
const emit = defineEmits(['update:modelValue'])

const items = [
  {
    title: 'Realtor Scraper',
    subtitle: 'Schedules & manual runs',
    icon: 'mdi-cog-transfer-outline',
    to: '/scraper',
  },
  {
    title: 'GTFS Data',
    subtitle: 'Transit feed refresh',
    icon: 'mdi-train',
    to: '/transit',
  },
]
</script>

<style scoped>
.settings-drawer__header {
  letter-spacing: 0.08em;
}
</style>
