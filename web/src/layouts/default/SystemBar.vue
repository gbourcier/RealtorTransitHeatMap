<template>
  <v-system-bar color="#262626" height="56" class="header-bar">
    <v-btn
      :to="'/listings'"
      :active="false"
      icon
      variant="text"
      size="small"
      class="ms-2"
      aria-label="Home"
    >
      <v-icon size="22">mdi-home-outline</v-icon>
    </v-btn>

    <v-btn
      icon
      variant="text"
      size="small"
      aria-label="Settings"
      @click="emit('click:settings')"
    >
      <v-icon size="22">mdi-cog-outline</v-icon>
    </v-btn>

    <div class="header-bar__divider" />

    <div id="header-filters-slot" class="header-bar__filters" />

    <v-spacer />

    <div id="header-count-slot" class="header-bar__count" />
  </v-system-bar>
</template>

<script lang="ts" setup>
const emit = defineEmits(['click:settings'])
</script>

<style scoped>
.header-bar :deep(.v-system-bar__content),
.header-bar.v-system-bar {
  padding-inline: 0;
}

.header-bar__divider {
  width: 1px;
  height: 24px;
  background-color: rgba(255, 255, 255, 0.16);
  margin: 0 8px 0 6px;
  flex: 0 0 auto;
}

.header-bar__filters {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  flex: 1 1 auto;
  overflow: hidden;
}

.header-bar__count {
  display: flex;
  align-items: center;
  font-size: 0.875rem;
  color: rgba(255, 255, 255, 0.65);
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
  margin-right: 12px;
  flex: 0 0 auto;
}
</style>
