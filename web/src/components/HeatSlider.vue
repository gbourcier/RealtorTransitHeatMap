<script setup lang="ts">
import { ref, computed } from "vue";

interface Props {
    min: number;
    max: number;
    step?: number;
    modelValue: number;
    heat?: boolean;
    ticks?: string[];
    ariaLabel?: string;
}

const props = withDefaults(defineProps<Props>(), {
    step: 1,
    heat: false,
    ticks: () => [],
    ariaLabel: undefined,
});

const emit = defineEmits<{ "update:modelValue": [value: number] }>();

const railRef = ref<HTMLElement | null>(null);

const fillPct = computed(() => {
    const ratio = (props.modelValue - props.min) / (props.max - props.min);
    return Math.max(0, Math.min(1, ratio)) * 100;
});

function valueFromClientX(clientX: number): number {
    const el = railRef.value;
    if (!el) return props.modelValue;
    const rect = el.getBoundingClientRect();
    let ratio = rect.width === 0 ? 0 : (clientX - rect.left) / rect.width;
    ratio = Math.max(0, Math.min(1, ratio));
    let v = props.min + ratio * (props.max - props.min);
    v = Math.round(v / props.step) * props.step;
    return Math.max(props.min, Math.min(props.max, v));
}

function commit(value: number): void {
    if (value !== props.modelValue) emit("update:modelValue", value);
}

function onPointerDown(e: PointerEvent): void {
    e.preventDefault();
    commit(valueFromClientX(e.clientX));
    const move = (ev: PointerEvent) => commit(valueFromClientX(ev.clientX));
    const up = () => {
        window.removeEventListener("pointermove", move);
        window.removeEventListener("pointerup", up);
    };
    window.addEventListener("pointermove", move);
    window.addEventListener("pointerup", up);
}

function onKeydown(e: KeyboardEvent): void {
    let v = props.modelValue;
    switch (e.key) {
        case "ArrowRight":
        case "ArrowUp":
            v += props.step;
            break;
        case "ArrowLeft":
        case "ArrowDown":
            v -= props.step;
            break;
        case "Home":
            v = props.min;
            break;
        case "End":
            v = props.max;
            break;
        default:
            return;
    }
    e.preventDefault();
    commit(Math.max(props.min, Math.min(props.max, v)));
}
</script>

<template>
    <div class="heat-slider">
        <div class="sld" @pointerdown="onPointerDown">
            <div ref="railRef" class="sld__rail" :class="{ 'sld__rail--heat': heat }" />
            <div
                v-if="!heat"
                class="sld__fill"
                :style="{ left: '0', right: 100 - fillPct + '%' }"
            />
            <div
                v-if="heat"
                class="sld__dim"
                :style="{ left: fillPct + '%', right: '0' }"
            />
            <div
                class="sld__knob"
                role="slider"
                tabindex="0"
                :aria-label="ariaLabel"
                :aria-valuemin="min"
                :aria-valuemax="max"
                :aria-valuenow="modelValue"
                :style="{ left: fillPct + '%' }"
                @pointerdown.stop="onPointerDown"
                @keydown="onKeydown"
            />
        </div>
        <div v-if="ticks.length" class="sld__ticks">
            <span v-for="(t, i) in ticks" :key="i">{{ t }}</span>
        </div>
    </div>
</template>

<style scoped>
.sld {
    position: relative;
    height: 22px;
    margin: 2px 2px 0;
    touch-action: none;
}

.sld__rail {
    position: absolute;
    top: 50%;
    left: 0;
    right: 0;
    height: 5px;
    transform: translateY(-50%);
    border-radius: 999px;
    background: rgba(var(--v-theme-on-surface), 0.14);
}

.sld__rail--heat {
    background: linear-gradient(
        90deg,
        rgb(var(--v-theme-commute-fast)) 0%,
        rgb(var(--v-theme-commute-mid)) 52%,
        rgb(var(--v-theme-commute-slow)) 100%
    );
}

.sld__fill {
    position: absolute;
    top: 50%;
    height: 5px;
    transform: translateY(-50%);
    border-radius: 999px;
    background: rgb(var(--v-theme-primary));
}

.sld__dim {
    position: absolute;
    top: 50%;
    height: 7px;
    transform: translateY(-50%);
    border-radius: 999px;
    background: rgb(var(--v-theme-surface));
    box-shadow: inset 0 0 0 1px rgba(var(--v-theme-on-surface), 0.06);
}

.sld__knob {
    position: absolute;
    top: 50%;
    width: 20px;
    height: 20px;
    transform: translate(-50%, -50%);
    border-radius: 50%;
    background: rgb(var(--v-theme-primary));
    border: 3px solid rgb(var(--v-theme-surface));
    box-shadow: 0 0 0 1px rgba(var(--v-theme-primary), 0.6),
        0 2px 5px rgba(var(--v-theme-shadow), 0.45);
    cursor: grab;
}

.sld__knob:active {
    cursor: grabbing;
}

.sld__knob:focus-visible {
    outline: 2px solid rgb(var(--v-theme-primary));
    outline-offset: 3px;
}

.sld__ticks {
    display: flex;
    justify-content: space-between;
    margin-top: 9px;
    font-size: 0.6875rem;
    color: rgba(var(--v-theme-on-surface), 0.42);
    font-variant-numeric: tabular-nums;
}
</style>
