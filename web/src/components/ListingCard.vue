<script setup lang="ts">
import type { Listing } from "../api/listings";
import {
    formatPrice,
    formatDate,
    isNew,
    formatCommute,
    commuteColor,
    commuteMapUrl,
    parseAddress,
} from "../utils/listingFormat";

interface Props {
    item: Listing;
    variant: "panel" | "mobile";
    selected?: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
    click: [item: Listing];
    hover: [item: Listing];
    leave: [];
}>();

function onActivate() {
    emit("click", props.item);
}
</script>

<template>
    <div
        :role="variant === 'panel' ? 'button' : 'link'"
        tabindex="0"
        class="listing-card"
        :class="{
            'listing-card--interactive': variant === 'panel',
            'listing-card--selected': variant === 'panel' && selected,
            'listing-card--mobile': variant === 'mobile',
        }"
        @click="onActivate"
        @keydown.enter.prevent="onActivate"
        @keydown.space.prevent="onActivate"
        @mouseenter="variant === 'panel' && emit('hover', item)"
        @mouseleave="variant === 'panel' && emit('leave')"
        @focus="variant === 'panel' && emit('hover', item)"
        @blur="variant === 'panel' && emit('leave')"
    >
        <div class="listing-card__top">
            <span class="listing-card__price">{{ formatPrice(item.currentPrice) }}</span>
            <v-chip
                v-if="isNew(item.firstSeenAt)"
                size="x-small"
                color="secondary"
                :variant="variant === 'panel' ? 'flat' : 'tonal'"
                class="listing-card__new"
            >new</v-chip>
        </div>
        <div class="listing-card__street">
            {{ parseAddress(item.address).street }}
        </div>
        <div v-if="parseAddress(item.address).locality" class="listing-card__locality">
            {{ parseAddress(item.address).locality }}
        </div>
        <div class="listing-card__meta">
            <a
                v-if="item.commuteSecondsDowntown != null && item.address"
                :href="commuteMapUrl(item.address) ?? '#'"
                target="_blank"
                rel="noopener noreferrer"
                class="listing-card__commute listing-card__commute--link"
                @click.stop
            >
                <span
                    class="listing-card__commute-dot"
                    :style="{ background: commuteColor(item.commuteSecondsDowntown) }"
                />
                <span>{{ formatCommute(item.commuteSecondsDowntown) }} downtown</span>
            </a>
            <span v-else class="listing-card__commute listing-card__commute--muted">
                <span
                    class="listing-card__commute-dot"
                    :style="{ background: commuteColor(null) }"
                />
                —
            </span>
            <span class="listing-card__seen">{{ formatDate(item.firstSeenAt) }}</span>
        </div>
    </div>
</template>

<style scoped>
.listing-card {
    position: relative;
    display: block;
    text-decoration: none;
    color: inherit;
    background-color: rgba(var(--v-theme-on-surface), 0.03);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    border-radius: 14px;
    padding: 14px 14px 12px;
    cursor: pointer;
    transition: background-color 120ms ease, border-color 120ms ease;
}

.listing-card:active {
    background-color: rgba(var(--v-theme-on-surface), 0.06);
}

.listing-card:focus-visible {
    outline: 2px solid rgb(var(--v-theme-secondary));
    outline-offset: 2px;
}

.listing-card--interactive:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.06);
    border-color: rgba(var(--v-theme-on-surface), 0.18);
}

.listing-card--selected,
.listing-card--selected:hover {
    background-color: rgba(var(--v-theme-secondary), 0.12);
    border-color: rgba(var(--v-theme-secondary), 0.55);
    box-shadow: inset 0 0 0 1px rgba(var(--v-theme-secondary), 0.45);
}

.listing-card--mobile {
    padding: 18px 18px 16px;
}

.listing-card--mobile .listing-card__price {
    font-size: 1.6rem;
    font-weight: 700;
}

.listing-card--mobile .listing-card__street {
    font-size: 1rem;
    font-weight: 500;
}

.listing-card--mobile .listing-card__meta {
    border-top: 0;
    padding-top: 14px;
    margin-top: 10px;
}

.listing-card--mobile .listing-card__new {
    top: 16px;
    right: 16px;
}

.listing-card__top {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 6px;
    padding-right: 56px;
}

.listing-card__price {
    font-size: 1.25rem;
    font-weight: 600;
    line-height: 1.2;
}

.listing-card__new {
    position: absolute;
    top: 12px;
    right: 12px;
    margin-left: 0;
}

.listing-card__street {
    font-size: 0.95rem;
    line-height: 1.3;
}

.listing-card__locality {
    font-size: 0.8125rem;
    color: rgba(var(--v-theme-on-surface), 0.65);
    margin-top: 2px;
}

.listing-card__meta {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin-top: 12px;
    padding-top: 10px;
    border-top: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    font-size: 0.8125rem;
}

.listing-card__commute {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    color: rgba(var(--v-theme-on-surface), 0.85);
    font-weight: 500;
    font-size: 0.8125rem;
}

.listing-card__commute--link {
    text-decoration: none;
    padding: 6px 12px;
    margin: -6px 0 -6px -4px;
    border-radius: 999px;
    background-color: rgba(var(--v-theme-on-surface), 0.06);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    transition: background-color 120ms ease, border-color 120ms ease;
}

.listing-card__commute--link:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.1);
    border-color: rgba(var(--v-theme-on-surface), 0.14);
}

.listing-card__commute--link:active {
    background-color: rgba(var(--v-theme-on-surface), 0.14);
}

.listing-card__commute--link:focus-visible {
    outline: 2px solid rgb(var(--v-theme-secondary));
    outline-offset: 2px;
}

.listing-card__commute--muted {
    color: rgba(var(--v-theme-on-surface), 0.5);
    font-weight: 400;
}

.listing-card__commute-dot {
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    flex-shrink: 0;
    box-shadow: 0 0 0 1px rgba(var(--v-theme-shadow), 0.3);
}

.listing-card__seen {
    margin-left: auto;
    color: rgba(var(--v-theme-on-surface), 0.45);
    font-size: 0.75rem;
}
</style>
