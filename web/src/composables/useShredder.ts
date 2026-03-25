import { ref } from 'vue'

export interface ShredderOptions {
  strips?: number
}

const STRIP_GRADIENTS = [
  'circle at 64% 0%',
  'circle at 109% 0%',
  'circle at 167% 173%',
  'circle at 118% -54%',
  'circle at 127% 57%',
  'circle at 116% -103%',
  'circle at 155% 143%',
  'circle at 241% 154%',
  'circle at 170% -219%',
  'circle at 217% 150%',
]

const ANIMATION_DURATION = 2750

export function useShredder(userOptions?: ShredderOptions) {
  const isShredding = ref(false)
  const stripCount = userOptions?.strips ?? 10

  function shred(container: HTMLElement): Promise<void> {
    if (isShredding.value) return Promise.resolve()
    isShredding.value = true

    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      return fadeOut(container)
    }

    return runShredAnimation(container)
  }

  function fadeOut(container: HTMLElement): Promise<void> {
    return new Promise((resolve) => {
      container.style.transition = 'opacity 300ms ease-out'
      container.style.opacity = '0'
      setTimeout(() => {
        isShredding.value = false
        resolve()
      }, 300)
    })
  }

  function runShredAnimation(container: HTMLElement): Promise<void> {
    return new Promise((resolve) => {
      const contentEl = container.firstElementChild as HTMLElement | null
      if (!contentEl) {
        isShredding.value = false
        resolve()
        return
      }

      const { width: cardWidth, height: cardHeight } = container.getBoundingClientRect()

      container.style.pointerEvents = 'none'
      container.style.zIndex = '9999'

      const feedClip = document.createElement('div')
      feedClip.className = 'shredder-feed-clip'
      feedClip.setAttribute('aria-hidden', 'true')
      feedClip.style.width = `${cardWidth}px`
      feedClip.style.height = `${cardHeight}px`

      const feedCard = contentEl.cloneNode(true) as HTMLElement
      feedCard.style.setProperty('--shredder-h', `${cardHeight}px`)
      feedClip.appendChild(feedCard)

      const shredderBar = document.createElement('div')
      shredderBar.className = 'shredder-bar'
      shredderBar.setAttribute('aria-hidden', 'true')
      shredderBar.style.top = `${cardHeight - 7}px`

      const stripClip = document.createElement('div')
      stripClip.className = 'shredder-strip-clip'
      stripClip.setAttribute('aria-hidden', 'true')
      stripClip.style.top = `${cardHeight + 7}px`
      stripClip.style.width = `${cardWidth}px`
      stripClip.style.height = `${cardHeight + 300}px`

      for (let i = 0; i < stripCount; i++) {
        const strip = document.createElement('div')
        strip.className = 'shredder-strip'
        strip.style.width = `${cardWidth}px`
        strip.style.height = `${cardHeight}px`

        const leftPct = (i / stripCount) * 100
        const rightPct = (1 - (i + 1) / stripCount) * 100
        strip.style.clipPath = `inset(0 ${rightPct}% 0 ${leftPct}%)`

        strip.appendChild(contentEl.cloneNode(true))

        const shadow = document.createElement('div')
        shadow.className = 'shredder-strip-shadow'
        shadow.style.background = `radial-gradient(${STRIP_GRADIENTS[i % STRIP_GRADIENTS.length]}, rgba(0,0,0,0.3), rgba(0,0,0,0) 70%)`
        strip.appendChild(shadow)

        strip.style.setProperty('--shredder-h', `${cardHeight}px`)
        stripClip.appendChild(strip)
      }

      container.appendChild(feedClip)
      container.appendChild(shredderBar)
      container.appendChild(stripClip)
      contentEl.style.opacity = '0'
      contentEl.style.transition = 'none'

      requestAnimationFrame(() => {
        requestAnimationFrame(() => {
          feedCard.classList.add('shredder-feed-card')
          shredderBar.classList.add('shredder-bar-shaking')
          stripClip.querySelectorAll('.shredder-strip').forEach((strip, i) => {
            strip.classList.add(i % 2 === 0 ? 'shredder-strip-even' : 'shredder-strip-odd')
          })
        })
      })

      setTimeout(() => {
        feedClip.remove()
        shredderBar.remove()
        stripClip.remove()
        container.style.zIndex = ''
        container.style.pointerEvents = ''
        isShredding.value = false
        resolve()
      }, ANIMATION_DURATION)
    })
  }

  return { isShredding, shred }
}
