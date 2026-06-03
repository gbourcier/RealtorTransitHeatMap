<script setup lang="ts">
withDefaults(
    defineProps<{
        text?: string;
        icon?: string;
    }>(),
    {
        text: "No listings found.",
        icon: "mdi-home-search-outline",
    },
);
</script>

<template>
    <div class="empty-state">
        <v-icon :icon="icon" size="40" class="empty-state__icon" />
        <p class="empty-state__text">{{ text }}</p>
    </div>
</template>

<style scoped>
.empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 48px 24px;
    text-align: center;
}

.empty-state__icon {
    color: rgba(var(--v-theme-on-surface), 0.35);
}

.empty-state__text {
    margin: 0;
    font-size: 0.9375rem;
    color: rgba(var(--v-theme-on-surface), 0.6);
}
</style>
