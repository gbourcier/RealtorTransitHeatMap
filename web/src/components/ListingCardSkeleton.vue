<script setup lang="ts">
defineProps<{ variant: "panel" | "mobile" }>();
</script>

<template>
    <div
        class="listing-card-skeleton"
        :class="{ 'listing-card-skeleton--mobile': variant === 'mobile' }"
        aria-hidden="true"
    >
        <div class="listing-card-skeleton__bar listing-card-skeleton__bar--price" />
        <div class="listing-card-skeleton__bar listing-card-skeleton__bar--street" />
        <div class="listing-card-skeleton__bar listing-card-skeleton__bar--locality" />
        <div class="listing-card-skeleton__meta">
            <div class="listing-card-skeleton__pill" />
            <div class="listing-card-skeleton__bar listing-card-skeleton__bar--seen" />
        </div>
    </div>
</template>

<style scoped>
.listing-card-skeleton {
    position: relative;
    background-color: rgba(var(--v-theme-on-surface), 0.03);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    border-radius: 14px;
    padding: 14px 14px 12px;
}

.listing-card-skeleton--mobile {
    padding: 18px 18px 16px;
}

.listing-card-skeleton__bar,
.listing-card-skeleton__pill {
    background: linear-gradient(
        90deg,
        rgba(var(--v-theme-on-surface), 0.04) 0%,
        rgba(var(--v-theme-on-surface), 0.12) 50%,
        rgba(var(--v-theme-on-surface), 0.04) 100%
    );
    background-size: 200% 100%;
    animation: listing-card-skeleton-shimmer 1.4s ease-in-out infinite;
    border-radius: 6px;
}

.listing-card-skeleton__bar--price {
    height: 22px;
    width: 42%;
    margin-bottom: 10px;
}

.listing-card-skeleton--mobile .listing-card-skeleton__bar--price {
    height: 28px;
    width: 48%;
}

.listing-card-skeleton__bar--street {
    height: 16px;
    width: 75%;
    margin-bottom: 6px;
}

.listing-card-skeleton--mobile .listing-card-skeleton__bar--street {
    height: 18px;
}

.listing-card-skeleton__bar--locality {
    height: 12px;
    width: 50%;
}

.listing-card-skeleton__meta {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin-top: 12px;
    padding-top: 10px;
    border-top: 1px solid rgba(var(--v-theme-on-surface), 0.08);
}

.listing-card-skeleton--mobile .listing-card-skeleton__meta {
    border-top: 0;
    padding-top: 14px;
    margin-top: 10px;
}

.listing-card-skeleton__pill {
    height: 28px;
    width: 140px;
    border-radius: 999px;
}

.listing-card-skeleton__bar--seen {
    height: 12px;
    width: 60px;
}

@keyframes listing-card-skeleton-shimmer {
    0% {
        background-position: 100% 0;
    }
    100% {
        background-position: -100% 0;
    }
}
</style>
