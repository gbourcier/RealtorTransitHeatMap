<script setup lang="ts">
interface Props {
    total: number;
    loading?: boolean;
}

withDefaults(defineProps<Props>(), {
    loading: false,
});
</script>

<template>
    <div class="count-pill" aria-live="polite">
        <span v-if="loading" class="count-pill__shimmer" aria-label="Loading listings" />
        <template v-else>
            <span class="count-pill__num">{{ total.toLocaleString() }}</span>
            <span class="count-pill__label">listing{{ total === 1 ? "" : "s" }}</span>
        </template>
    </div>
</template>

<style scoped>
.count-pill {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 5px;
    height: 40px;
    padding: 0;
    background: transparent;
    color: #f4f1e8;
    font-family: Inter, system-ui, sans-serif;
    font-size: 13.5px;
    font-weight: 500;
    letter-spacing: 0;
    line-height: 1;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
}

.count-pill__num {
    font-size: 14px;
    font-weight: 700;
}

.count-pill__label {
    font-weight: 500;
    color: rgba(244, 241, 232, 0.52);
}

.count-pill__shimmer {
    display: inline-block;
    width: 82px;
    height: 13px;
    border-radius: 7px;
    background: linear-gradient(90deg, #2a2d27 25%, #3a3e36 50%, #2a2d27 75%);
    background-size: 200% 100%;
    animation: count-shimmer 1.1s linear infinite;
}

@keyframes count-shimmer {
    to {
        background-position: -200% 0;
    }
}

@media (max-width: 899px) {
    .count-pill {
        flex-direction: column;
        align-items: center;
        gap: 2px;
        height: 38px;
        min-width: 44px;
        line-height: 1;
    }

    .count-pill__label {
        display: block;
        font-size: 10px;
        font-weight: 600;
    }

    .count-pill__shimmer {
        width: 44px;
    }
}
</style>
