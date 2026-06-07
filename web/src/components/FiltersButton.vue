<script setup lang="ts">
interface Props {
    open: boolean;
    count: number;
}

defineProps<Props>();
const emit = defineEmits<{ toggle: [] }>();
</script>

<template>
    <button
        type="button"
        class="filters-btn"
        :class="{ 'filters-btn--on': open }"
        :aria-pressed="open"
        aria-label="Toggle filters"
        @click="emit('toggle')"
    >
        <v-icon size="15" class="filters-btn__icon">mdi-tune-variant</v-icon>
        <span class="filters-btn__label">Filters</span>
        <span v-if="count > 0" class="filters-btn__badge">{{ count }}</span>
    </button>
</template>

<style scoped>
.filters-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 28px;
    padding: 0 12px;
    border-radius: 999px;
    white-space: nowrap;
    background: rgba(var(--v-theme-on-surface), 0.06);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.14);
    color: rgba(var(--v-theme-on-surface), 0.92);
    font-size: 0.8125rem;
    font-weight: 600;
    letter-spacing: normal;
    cursor: pointer;
    transition: background-color 120ms ease, border-color 120ms ease, color 120ms ease;
}

.filters-btn:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.1);
}

.filters-btn:focus-visible {
    outline: 2px solid rgb(var(--v-theme-primary));
    outline-offset: 2px;
}

.filters-btn--on,
.filters-btn--on:hover {
    background-color: rgba(var(--v-theme-primary), 0.14);
    border-color: rgba(var(--v-theme-primary), 0.5);
    color: rgb(var(--v-theme-primary));
}

.filters-btn__icon {
    opacity: 0.95;
}

.filters-btn__badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 18px;
    height: 18px;
    padding: 0 5px;
    border-radius: 999px;
    background: rgb(var(--v-theme-primary));
    color: rgb(var(--v-theme-on-primary));
    font-size: 0.6875rem;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
}
</style>
