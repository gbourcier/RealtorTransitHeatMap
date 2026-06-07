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
        :class="{ 'filters-btn--on': open, 'filters-btn--active': count > 0 }"
        :aria-pressed="open"
        aria-label="Toggle filters"
        @click="emit('toggle')"
    >
        <v-icon size="18" class="filters-btn__icon">mdi-tune-variant</v-icon>
        <span class="filters-btn__label">Filters</span>
        <span v-if="count > 0" class="filters-btn__badge">{{ count }}</span>
    </button>
</template>

<style scoped>
.filters-btn {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    height: 40px;
    padding: 0 14px;
    border-radius: 999px;
    white-space: nowrap;
    background: #2a2d27;
    border: 1px solid rgba(244, 241, 232, 0.12);
    color: #f4f1e8;
    box-shadow: none;
    font-family: Inter, system-ui, sans-serif;
    font-size: 14px;
    font-weight: 700;
    letter-spacing: 0;
    cursor: pointer;
    transition: background-color 140ms ease, border-color 140ms ease, box-shadow 140ms ease, color 140ms ease, transform 60ms ease;
}

.filters-btn:hover {
    background-color: #34382f;
    border-color: rgba(244, 241, 232, 0.22);
}

.filters-btn:active {
    transform: translateY(1px);
}

.filters-btn:focus-visible {
    outline: 2px solid #6ccff6;
    outline-offset: 2px;
}

.filters-btn--on,
.filters-btn--on:hover {
    background-color: #34382f;
    border-color: rgba(244, 241, 232, 0.22);
}

.filters-btn--active {
    background: #b6f24a;
    border-color: #b6f24a;
    color: #172006;
    box-shadow: 0 2px 10px -2px rgba(182, 242, 74, 0.4);
}

.filters-btn--active:hover,
.filters-btn--active.filters-btn--on {
    background-color: #c6fb60;
    border-color: #c6fb60;
    box-shadow: 0 4px 14px -2px rgba(182, 242, 74, 0.5);
}

.filters-btn__icon {
    opacity: 0.95;
}

.filters-btn__badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 20px;
    height: 20px;
    padding: 0 5px;
    border-radius: 999px;
    background: #172006;
    color: #b6f24a;
    font-size: 12px;
    font-weight: 800;
    font-variant-numeric: tabular-nums;
}

@media (max-width: 899px) {
    .filters-btn {
        height: 38px;
        padding: 0 13px;
        gap: 11px;
    }

    .filters-btn__label {
        display: none;
    }
}
</style>
