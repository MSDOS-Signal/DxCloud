<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { zhCN, dateZhCN, darkTheme } from 'naive-ui'
import type { GlobalThemeOverrides } from 'naive-ui'

const theme = useThemeStore()

// 浅色主题：腾讯云蓝主色 + 小圆角 + 简洁卡片
const lightOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#006EFF',
    primaryColorHover: '#005CE6',
    primaryColorPressed: '#0050CC',
    primaryColorSuppl: '#3D8BFF',
    infoColor: '#00A4FF',
    successColor: '#00B42A',
    warningColor: '#FF9500',
    errorColor: '#F53F3F',
    borderRadius: '4px',
    fontSize: '14px',
    fontFamily: `'PingFang SC', 'HarmonyOS Sans SC', 'Microsoft YaHei', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif`,
  },
  Card: {
    borderRadius: '4px',
    boxShadow: '0 1px 2px rgba(0, 0, 0, 0.04)',
    borderColor: '#E8E8E8',
    titleFontSizeMedium: '14px',
    titleFontWeight: '600',
  },
  Button: {
    borderRadiusMedium: '4px',
    borderRadiusLarge: '4px',
    fontWeight: '400',
  },
  Input: {
    borderRadius: '4px',
  },
  DataTable: {
    thColor: '#FAFAFA',
    thFontWeight: '500',
    thTextColor: '#595959',
    borderColor: '#E8E8E8',
  },
  Menu: {
    itemHeight: '36px',
    borderRadius: '4px',
    itemColorActive: '#E8F3FF',
    itemColorActiveHover: '#D4E8FF',
    itemTextColorActive: '#006EFF',
    itemIconColorActive: '#006EFF',
    itemTextColor: '#595959',
    itemIconColor: '#8C8C8C',
    itemTextColorHover: '#262626',
    itemIconColorHover: '#595959',
    itemTextColorActiveHover: '#005CE6',
    itemIconColorActiveHover: '#005CE6',
    arrowColor: '#8C8C8C',
    groupTextColor: '#8C8C8C',
  },
  Tag: {
    borderRadius: '2px',
  },
  Modal: {
    borderRadius: '8px',
  },
}

// 深色主题
const darkOverrides: GlobalThemeOverrides = {
  common: {
    primaryColor: '#006EFF',
    primaryColorHover: '#3D8BFF',
    primaryColorPressed: '#0050CC',
    primaryColorSuppl: '#3D8BFF',
    infoColor: '#00A4FF',
    successColor: '#23C343',
    warningColor: '#FF9500',
    errorColor: '#F53F3F',
    borderRadius: '4px',
    fontSize: '14px',
    fontFamily: `'PingFang SC', 'HarmonyOS Sans SC', 'Microsoft YaHei', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif`,
    bodyColor: '#0D1117',
    cardColor: '#161B22',
    modalColor: '#161B22',
    popoverColor: '#161B22',
    tableColor: '#161B22',
    tableHeaderColor: '#1C2128',
    inputColor: '#0D1117',
    borderColor: '#30363D',
    dividerColor: '#30363D',
  },
  Card: {
    borderRadius: '4px',
    borderColor: '#30363D',
    titleFontSizeMedium: '14px',
    titleFontWeight: '600',
  },
  Button: {
    borderRadiusMedium: '4px',
    borderRadiusLarge: '4px',
    fontWeight: '400',
  },
  Input: {
    borderRadius: '4px',
  },
  DataTable: {
    thColor: '#1C2128',
    thFontWeight: '500',
    thTextColor: '#8B949E',
    borderColor: '#30363D',
  },
  Tag: {
    borderRadius: '2px',
  },
  Modal: {
    borderRadius: '8px',
  },
}

const themeOverrides = computed(() => (theme.isDark ? darkOverrides : lightOverrides))

onMounted(() => {
  theme.apply()
})
</script>

<template>
  <n-config-provider
    :locale="zhCN"
    :date-locale="dateZhCN"
    :theme="theme.isDark ? darkTheme : null"
    :theme-overrides="themeOverrides"
  >
    <n-global-style />
    <n-message-provider>
      <n-dialog-provider>
        <n-notification-provider>
          <NuxtLayout>
            <NuxtPage :transition="{ name: 'page', mode: 'out-in', appear: true }" />
          </NuxtLayout>
        </n-notification-provider>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>
