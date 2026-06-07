<script setup lang="ts">
import { ref, computed } from "vue";

interface Tick {
    at: number;
    label: string;
}

interface Props {
    min: number;
    max: number;
    step?: number;
    modelValue: number;
    heat?: boolean;
    ticks?: Tick[];
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

function tickPct(at: number): number {
    const ratio = (at - props.min) / (props.max - props.min);
    return Math.max(0, Math.min(1, ratio)) * 100;
}

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
            <span
                v-for="(t, i) in ticks"
                :key="i"
                :style="{ left: tickPct(t.at) + '%', transform: `translateX(-${tickPct(t.at)}%)` }"
            >{{ t.label }}</span>
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
    height: 6px;
    transform: translateY(-50%);
    border-radius: 999px;
    background: rgba(244, 241, 232, 0.14);
}

.sld__rail--heat {
    background: linear-gradient(
        90deg,
        #9be84a 0%,
        #ffb454 55%,
        #ff5c8a 100%
    );
}

.sld__fill {
    position: absolute;
    top: 50%;
    height: 6px;
    transform: translateY(-50%);
    border-radius: 999px;
    background: #b6f24a;
}

.sld__dim {
    position: absolute;
    top: 50%;
    height: 6px;
    transform: translateY(-50%);
    border-radius: 999px;
    background: #2a2d27;
    box-shadow: inset 0 0 0 1px rgba(244, 241, 232, 0.06);
}

.sld__knob {
    position: absolute;
    top: 50%;
    width: 20px;
    height: 20px;
    transform: translate(-50%, -50%);
    border-radius: 50%;
    background: #b6f24a;
    border: 3px solid #181a17;
    box-shadow: 0 0 0 1.5px #b6f24a, 0 2px 6px rgba(0, 0, 0, 0.5);
    cursor: grab;
}

.sld__knob:active {
    cursor: grabbing;
}

.sld__knob:focus-visible {
    outline: 2px solid #6ccff6;
    outline-offset: 3px;
}

.sld__ticks {
    position: relative;
    height: 0.85rem;
    margin-top: 9px;
    font-size: 12px;
    color: rgba(244, 241, 232, 0.52);
    font-variant-numeric: tabular-nums;
}

.sld__ticks span {
    position: absolute;
    top: 0;
    white-space: nowrap;
}
</style>
