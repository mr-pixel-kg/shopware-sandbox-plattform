<script setup lang="ts">
import { ref } from 'vue'

import { useShredder } from '@/composables/useShredder'

const props = withDefaults(defineProps<{ strips?: number }>(), {
  strips: 10,
})

const wrapperRef = ref<HTMLElement>()
const { isShredding, shred: shredEl } = useShredder(props)

async function shred(): Promise<void> {
  if (!wrapperRef.value) return
  await shredEl(wrapperRef.value)
}

defineExpose({ shred, isShredding })
</script>

<template>
  <div ref="wrapperRef" class="shredder-wrapper">
    <slot />
  </div>
</template>

<style>
.shredder-wrapper {
  position: relative;
  overflow: visible;
}

.shredder-feed-clip {
  position: absolute;
  top: 0;
  left: 0;
  overflow: hidden;
  z-index: 10;
  pointer-events: none;
}

.shredder-feed-card {
  position: relative;
  top: 0;
  animation: shredder-feed-down 2.75s 1 linear forwards;
}

.shredder-bar {
  position: absolute;
  left: -4px;
  right: -4px;
  height: 14px;
  border-radius: 3px;
  z-index: 9999999;
  pointer-events: none;
  background: linear-gradient(
    to bottom,
    hsl(0 0% 62%) 0%,
    hsl(0 0% 75%) 35%,
    hsl(0 0% 70%) 50%,
    hsl(0 0% 58%) 100%
  );
  box-shadow:
    0 2px 8px rgba(0, 0, 0, 0.25),
    inset 0 1px 0 rgba(255, 255, 255, 0.3);
}

.shredder-bar-shaking {
  animation: shredder-shake 0.06s 26 0.25s;
}

.shredder-strip-clip {
  position: absolute;
  left: 0;
  overflow: hidden;
  pointer-events: none;
  perspective: 1000px;
  z-index: 5;
}

.shredder-strip {
  position: absolute;
  left: 0;
  transform-style: preserve-3d;
  will-change: transform;
  z-index: 99;
  overflow: hidden;
  border-radius: var(--radius-xl, 0.75rem);
}

.shredder-strip-shadow {
  position: absolute;
  inset: 0;
  pointer-events: none;
  border-radius: inherit;
}

.shredder-strip-even {
  animation: shredder-strip-even 2.75s 1 linear forwards;
}

.shredder-strip-odd {
  animation: shredder-strip-odd 2.75s 1 linear forwards;
}

@keyframes shredder-feed-down {
  0% {
    top: 0;
  }
  7% {
    top: 0;
  }
  20% {
    top: calc(var(--shredder-h) * 0.3);
  }
  35% {
    top: calc(var(--shredder-h) * 0.55);
  }
  50% {
    top: calc(var(--shredder-h) * 0.8);
  }
  60% {
    top: calc(var(--shredder-h) * 0.95);
  }
  64% {
    top: var(--shredder-h);
  }
  100% {
    top: var(--shredder-h);
  }
}

@keyframes shredder-shake {
  0% {
    transform: translateX(0);
  }
  50% {
    transform: translateX(0.8px);
  }
  100% {
    transform: translateX(-0.8px);
  }
}

@keyframes shredder-strip-even {
  0% {
    top: calc(var(--shredder-h) * -1);
    transform: scaleY(1) rotateX(0deg);
    opacity: 1;
  }
  7% {
    top: calc(var(--shredder-h) * -1);
    transform: scaleY(1) rotateX(0deg);
    opacity: 1;
  }
  20% {
    top: calc(var(--shredder-h) * -0.7);
    transform: scaleY(1.02) rotateX(2deg);
    opacity: 1;
  }
  35% {
    top: calc(var(--shredder-h) * -0.45);
    transform: scaleY(1.04) rotateX(3deg);
    opacity: 1;
  }
  50% {
    top: calc(var(--shredder-h) * -0.2);
    transform: scaleY(1.06) rotateX(5deg);
    opacity: 1;
  }
  60% {
    top: calc(var(--shredder-h) * -0.05);
    transform: scaleY(1.08) rotateX(6deg);
    opacity: 1;
  }
  64% {
    top: 0;
    transform: scaleY(1.1) rotateX(7deg);
    opacity: 1;
  }
  76% {
    top: 150px;
    transform: scaleY(1.1) rotateX(9deg);
    opacity: 0;
  }
  100% {
    top: 150px;
    transform: scaleY(1.1) rotateX(9deg);
    opacity: 0;
  }
}

@keyframes shredder-strip-odd {
  0% {
    top: calc(var(--shredder-h) * -1);
    transform: scaleY(1) rotateX(0deg);
    opacity: 1;
  }
  7% {
    top: calc(var(--shredder-h) * -1);
    transform: scaleY(1) rotateX(0deg);
    opacity: 1;
  }
  20% {
    top: calc(var(--shredder-h) * -0.7);
    transform: scaleY(1.02) rotateX(-2deg);
    opacity: 1;
  }
  35% {
    top: calc(var(--shredder-h) * -0.45);
    transform: scaleY(1.04) rotateX(-3deg);
    opacity: 1;
  }
  50% {
    top: calc(var(--shredder-h) * -0.2);
    transform: scaleY(1.06) rotateX(-5deg);
    opacity: 1;
  }
  60% {
    top: calc(var(--shredder-h) * -0.05);
    transform: scaleY(1.08) rotateX(-6deg);
    opacity: 1;
  }
  64% {
    top: 0;
    transform: scaleY(1.1) rotateX(-7deg);
    opacity: 1;
  }
  76% {
    top: 150px;
    transform: scaleY(1.1) rotateX(-9deg);
    opacity: 0;
  }
  100% {
    top: 150px;
    transform: scaleY(1.1) rotateX(-9deg);
    opacity: 0;
  }
}
</style>
