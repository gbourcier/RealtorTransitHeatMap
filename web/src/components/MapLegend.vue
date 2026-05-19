<template>
    <div class="mobile-legend" aria-label="Transit commute time legend">
        <span class="mobile-legend__item">
            <span class="mobile-legend__dot mobile-legend__dot--fast" />
            &lt; 30
        </span>
        <span class="mobile-legend__item">
            <span class="mobile-legend__dot mobile-legend__dot--mid" />
            30–60
        </span>
        <span class="mobile-legend__item">
            <span class="mobile-legend__dot mobile-legend__dot--slow" />
            &gt; 60 min
        </span>
    </div>
</template>

<style scoped>
.mobile-legend {
    position: fixed;
    left: 12px;
    bottom: calc(20px + env(safe-area-inset-bottom, 0px));
    z-index: 1000;
    display: inline-flex;
    align-items: center;
    gap: 14px;
    height: 40px;
    padding: 0 16px;
    border-radius: 999px;
    background-color: rgba(var(--v-theme-surface), 0.78);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    box-shadow: 0 6px 18px rgba(var(--v-theme-shadow), 0.45);
    color: rgba(var(--v-theme-on-surface), 0.92);
    font-size: 0.75rem;
    white-space: nowrap;
}

.mobile-legend__item {
    display: inline-flex;
    align-items: center;
    gap: 6px;
}

.mobile-legend__dot {
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    box-shadow: 0 0 0 1px rgba(var(--v-theme-shadow), 0.3);
}

.mobile-legend__dot--fast {
    background-color: rgb(var(--v-theme-success));
}

.mobile-legend__dot--mid {
    background-color: rgb(var(--v-theme-warning));
}

.mobile-legend__dot--slow {
    background-color: rgb(var(--v-theme-error));
}
</style>
