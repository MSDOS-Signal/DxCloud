// ECharts 按需引入：拆成小模块文件，避免 dev 模式下单个超大 chunk 加载中断（ERR_ABORTED）
import * as echarts from 'echarts/core'
import { LineChart, BarChart, PieChart, GaugeChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent,
  DataZoomComponent,
  MarkLineComponent,
  ToolboxComponent,
  GraphicComponent,
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import type { Ref } from 'vue'
import { nextTick } from 'vue'

echarts.use([
  LineChart,
  BarChart,
  PieChart,
  GaugeChart,
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent,
  DataZoomComponent,
  MarkLineComponent,
  ToolboxComponent,
  GraphicComponent,
  CanvasRenderer,
])

export type { EChartsOption } from 'echarts/core'
export { echarts }

/**
 * render() 的可选开关。
 * 全部默认关闭，保证现有 render(option) 调用行为完全不变。
 */
export interface RenderOptions {
  /** 在 setOption 前展示 ECharts 内置 loading 动画，option 生效后自动隐藏 */
  loading?: boolean
  /** loading 提示文字（默认无文字，仅转圈） */
  loadingText?: string
}

// ECharts 组合式封装：init/option 更新/resize/销毁
export function useEcharts(elRef: Ref<HTMLElement | null>) {
  let chart: echarts.ECharts | null = null
  let observer: ResizeObserver | null = null

  async function render(option: echarts.EChartsCoreOption, opts: RenderOptions = {}) {
    if (!elRef.value) {
      console.warn('[useEcharts] elRef is null')
      return
    }

    // 等待 DOM 完成布局，确保容器有尺寸
    await nextTick()

    if (!elRef.value) {
      console.warn('[useEcharts] elRef is null after nextTick')
      return
    }

    // 检查容器尺寸，如果为 0 则等待（最多 1s）
    let retries = 0
    while ((!elRef.value.offsetWidth || !elRef.value.offsetHeight) && retries < 20) {
      await new Promise((r) => setTimeout(r, 50))
      retries++
    }

    if (!elRef.value || !elRef.value.offsetWidth || !elRef.value.offsetHeight) {
      console.warn('[useEcharts] container has no dimensions', {
        w: elRef.value?.offsetWidth,
        h: elRef.value?.offsetHeight,
      })
      return
    }

    try {
      if (!chart || chart.isDisposed()) {
        chart = echarts.init(elRef.value)
        observer?.disconnect()
        observer = new ResizeObserver(() => chart?.resize())
        observer.observe(elRef.value)
      }
      // 可选 loading 开关：默认关闭，不改变现有调用行为
      if (opts.loading) {
        chart.showLoading('default', {
          text: opts.loadingText ?? '',
          color: '#006eff',
          textColor: '#86909c',
          maskColor: 'transparent',
          spinnerRadius: 12,
          lineWidth: 2,
          fontSize: 12,
        })
      }
      chart.setOption(option, { notMerge: true })
      if (opts.loading) {
        chart.hideLoading()
      }
    } catch (err) {
      console.error('[useEcharts] init/setOption failed:', err)
    }
  }

  function resize() {
    chart?.resize()
  }

  function dispose() {
    observer?.disconnect()
    observer = null
    chart?.dispose()
    chart = null
  }

  return { render, dispose, resize, getChart: () => chart }
}
