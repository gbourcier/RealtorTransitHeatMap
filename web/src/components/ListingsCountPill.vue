<script setup lang="ts">
interface Props {
    total: number;
}

defineProps<Props>();
</script>

<template>
    <div class="count-pill" aria-live="polite">
        <span class="count-pill__num">{{ total.toLocaleString() }}</span>
        <span class="count-pill__label">listing{{ total === 1 ? "" : "s" }}</span>
    </div>
</template>

<style scoped>
.count-pill {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 5px;
    height: 28px;
    padding: 0 10px;
    border-radius: 999px;
    background: transparent;
    color: rgba(var(--v-theme-on-surface), 0.88);
    font-size: 0.75rem;
    font-weight: 500;
    letter-spacing: normal;
    line-height: 1;
    font-variant-numeric: tabular-nums;
}

.count-pill__num {
    font-weight: 500;
}

.count-pill__label {
    font-weight: 500;
}
</style>
