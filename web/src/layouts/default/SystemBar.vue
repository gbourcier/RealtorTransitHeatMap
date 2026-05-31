<template>
  <v-system-bar color="header-bar" height="48" class="header-bar">
    <button
      v-if="onSettingsPage"
      type="button"
      class="back-pill ms-2"
      aria-label="Back to map"
      @click="emit('click:back')"
    >
      <v-icon size="16" class="back-pill__icon">mdi-arrow-left</v-icon>
    </button>

    <button
      v-if="authStore.isAdmin"
      type="button"
      class="settings-pill"
      :class="[
        onSettingsPage ? 'ms-1' : 'ms-2',
        { 'settings-pill--active': settingsActive },
      ]"
      :aria-pressed="settingsActive"
      aria-label="Settings"
      @click="emit('click:settings')"
    >
      <v-icon size="14" class="settings-pill__icon">mdi-cog-outline</v-icon>
      <span class="settings-pill__label">Settings</span>
    </button>

    <v-spacer />

    <div id="header-filters-slot" class="header-bar__filters" />
    <div id="header-actions-slot" class="header-bar__actions" />

    <v-menu location="bottom end" :close-on-content-click="true">
      <template #activator="{ props: menuProps }">
        <button type="button" class="user-pill me-2" v-bind="menuProps" aria-label="User menu">
          <v-icon size="14" class="user-pill__icon">mdi-account-outline</v-icon>
          <span class="user-pill__label">{{ authStore.user?.username }}</span>
          <v-icon size="12" class="user-pill__chevron">mdi-chevron-down</v-icon>
        </button>
      </template>
      <v-list density="compact" min-width="140">
        <v-list-item
          prepend-icon="mdi-logout"
          title="Sign out"
          @click="onLogout"
        />
      </v-list>
    </v-menu>
  </v-system-bar>
</template>

<script lang="ts" setup>
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../stores/auth'

withDefaults(defineProps<{ settingsActive?: boolean; onSettingsPage?: boolean }>(), {
  settingsActive: false,
  onSettingsPage: false,
})
const emit = defineEmits(['click:settings', 'click:back'])

const authStore = useAuthStore()
const router = useRouter()

async function onLogout() {
  await authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.header-bar :deep(.v-system-bar__content),
.header-bar.v-system-bar {
  padding-inline: 0;
}

.header-bar__filters {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  flex: 0 0 auto;
}

.header-bar__actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex: 0 0 auto;
}

.header-bar__actions:not(:empty) {
  margin-left: 6px;
}

.back-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 28px;
  width: 28px;
  padding: 0;
  border-radius: 999px;
  border: 1px solid rgba(var(--v-theme-secondary), 0.7);
  background: transparent;
  color: rgb(var(--v-theme-secondary));
  cursor: pointer;
  transition: background-color 120ms ease, border-color 120ms ease, color 120ms ease;
}

.back-pill:hover {
  background-color: rgba(var(--v-theme-secondary), 0.08);
  border-color: rgb(var(--v-theme-secondary));
}

.back-pill:focus-visible {
  outline: 2px solid rgb(var(--v-theme-secondary));
  outline-offset: 2px;
}

.back-pill__icon {
  opacity: 0.95;
}

.settings-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 28px;
  padding: 0 10px;
  border-radius: 999px;
  border: 1px solid rgba(var(--v-theme-on-surface), 0.22);
  background: transparent;
  color: rgba(var(--v-theme-on-surface), 0.88);
  font-size: 0.75rem;
  font-weight: 500;
  letter-spacing: normal;
  cursor: pointer;
  transition: background-color 120ms ease, border-color 120ms ease, color 120ms ease;
}

.settings-pill:hover {
  background-color: rgba(var(--v-theme-on-surface), 0.05);
  border-color: rgba(var(--v-theme-on-surface), 0.32);
}

.settings-pill:focus-visible {
  outline: 2px solid rgb(var(--v-theme-secondary));
  outline-offset: 2px;
}

.settings-pill--active {
  background-color: rgba(var(--v-theme-on-surface), 0.12);
  border-color: rgba(var(--v-theme-on-surface), 0.45);
  color: rgb(var(--v-theme-on-surface));
}

.settings-pill--active:hover {
  background-color: rgba(var(--v-theme-on-surface), 0.18);
  border-color: rgba(var(--v-theme-on-surface), 0.6);
}

.settings-pill__icon {
  opacity: 0.9;
}

.user-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 28px;
  padding: 0 10px;
  border-radius: 999px;
  border: 1px solid rgba(var(--v-theme-on-surface), 0.22);
  background: transparent;
  color: rgba(var(--v-theme-on-surface), 0.88);
  font-size: 0.75rem;
  font-weight: 500;
  letter-spacing: normal;
  cursor: pointer;
  transition: background-color 120ms ease, border-color 120ms ease;
}

.user-pill:hover {
  background-color: rgba(var(--v-theme-on-surface), 0.05);
  border-color: rgba(var(--v-theme-on-surface), 0.32);
}

.user-pill:focus-visible {
  outline: 2px solid rgb(var(--v-theme-secondary));
  outline-offset: 2px;
}

.user-pill__icon {
  opacity: 0.9;
}

.user-pill__chevron {
  opacity: 0.6;
}
</style>
