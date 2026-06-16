<script setup lang="ts">
defineProps<{ variant: "panel" | "mobile" }>();
</script>

<template>
    <div
        class="listing-card-skeleton"
        :class="{ 'listing-card-skeleton--mobile': variant === 'mobile' }"
        aria-hidden="true"
    >
        <div class="listing-card-skeleton__content">
            <div class="listing-card-skeleton__thumb" />
            <div class="listing-card-skeleton__details">
                <div class="listing-card-skeleton__bar listing-card-skeleton__bar--price" />
                <div class="listing-card-skeleton__bar listing-card-skeleton__bar--street" />
                <div class="listing-card-skeleton__bar listing-card-skeleton__bar--locality" />
                <div class="listing-card-skeleton__meta">
                    <div class="listing-card-skeleton__pill" />
                    <div class="listing-card-skeleton__bar listing-card-skeleton__bar--seen" />
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.listing-card-skeleton {
    position: relative;
    background-color: rgba(var(--v-theme-on-surface), 0.03);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    border-radius: 8px;
    padding: 0;
    overflow: hidden;
}

.listing-card-skeleton--mobile {
    padding: 0;
}

.listing-card-skeleton__bar,
.listing-card-skeleton__pill,
.listing-card-skeleton__thumb {
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

.listing-card-skeleton__content {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    min-width: 0;
}

.listing-card-skeleton__thumb {
    flex: 0 0 auto;
    width: 100%;
    height: 136px;
    border-radius: 0;
}

.listing-card-skeleton--mobile .listing-card-skeleton__thumb {
    height: 168px;
}

.listing-card-skeleton__details {
    flex: 1 1 auto;
    min-width: 0;
    padding: 14px;
}

.listing-card-skeleton__bar--price {
    height: 25px;
    width: 48%;
    margin-bottom: 10px;
}

.listing-card-skeleton--mobile .listing-card-skeleton__bar--price {
    height: 28px;
    width: 48%;
}

.listing-card-skeleton__bar--street {
    height: 17px;
    width: 82%;
    margin-bottom: 6px;
}

.listing-card-skeleton--mobile .listing-card-skeleton__bar--street {
    height: 18px;
}

.listing-card-skeleton__bar--locality {
    height: 14px;
    width: 58%;
}

.listing-card-skeleton__meta {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin-top: 16px;
}

.listing-card-skeleton--mobile .listing-card-skeleton__meta {
    margin-top: 16px;
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
