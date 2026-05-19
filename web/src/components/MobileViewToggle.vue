<script setup lang="ts">
interface Props {
    modelValue: "list" | "map";
}

defineProps<Props>();

const emit = defineEmits<{
    "update:modelValue": [value: "list" | "map"];
}>();
</script>

<template>
    <div class="mobile-view-toggle" role="tablist" aria-label="View mode">
        <button
            type="button"
            class="mobile-view-toggle__btn"
            :class="{ 'mobile-view-toggle__btn--active': modelValue === 'map' }"
            role="tab"
            :aria-selected="modelValue === 'map'"
            aria-label="Map view"
            @click="emit('update:modelValue', 'map')"
        >
            <v-icon size="20">mdi-map-outline</v-icon>
        </button>
        <button
            type="button"
            class="mobile-view-toggle__btn"
            :class="{ 'mobile-view-toggle__btn--active': modelValue === 'list' }"
            role="tab"
            :aria-selected="modelValue === 'list'"
            aria-label="List view"
            @click="emit('update:modelValue', 'list')"
        >
            <v-icon size="20">mdi-format-list-bulleted</v-icon>
        </button>
    </div>
</template>

<style scoped>
.mobile-view-toggle {
    position: fixed;
    right: 12px;
    bottom: calc(20px + env(safe-area-inset-bottom, 0px));
    z-index: 1000;
    display: inline-flex;
    align-items: center;
    height: 40px;
    padding: 4px;
    border-radius: 999px;
    background-color: rgba(var(--v-theme-surface), 0.78);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    box-shadow: 0 6px 18px rgba(var(--v-theme-shadow), 0.45);
}

.mobile-view-toggle__btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 44px;
    height: 32px;
    padding: 0;
    background: transparent;
    border: 0;
    border-radius: 999px;
    color: rgba(var(--v-theme-on-surface), 0.7);
    cursor: pointer;
    transition: background-color 120ms ease, color 120ms ease;
}

.mobile-view-toggle__btn:hover {
    color: rgba(var(--v-theme-on-surface), 0.95);
}

.mobile-view-toggle__btn:focus-visible {
    outline: 2px solid rgb(var(--v-theme-secondary));
    outline-offset: 2px;
}

.mobile-view-toggle__btn--active {
    background-color: rgba(var(--v-theme-on-surface), 0.14);
    color: rgb(var(--v-theme-on-surface));
}
</style>
