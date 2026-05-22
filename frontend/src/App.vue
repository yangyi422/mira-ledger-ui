<script setup>
import { ref } from 'vue'
import { GetVerify, GetMonthlyStats, RunGitHubBackup } from '../wailsjs/go/main/App'

const month = ref('2026-05')
const verifyOutput = ref('')
const statsOutput = ref('')
const backupOutput = ref('')
const loading = ref('')

async function runVerify() {
  loading.value = 'verify'
  verifyOutput.value = await GetVerify()
  loading.value = ''
}

async function loadStats() {
  loading.value = 'stats'
  statsOutput.value = await GetMonthlyStats(month.value)
  loading.value = ''
}

async function runBackup() {
  loading.value = 'backup'
  backupOutput.value = await RunGitHubBackup(month.value)
  loading.value = ''
}
</script>

<template>
  <main class="dashboard">
    <h1>MiraLedger</h1>

    <section class="controls">
      <label>
        月份
        <input v-model="month" type="month" />
      </label>

      <button type="button" :disabled="loading === 'stats'" @click="loadStats">
        加载统计
      </button>
      <button type="button" :disabled="loading === 'verify'" @click="runVerify">
        验证数据库
      </button>
      <button type="button" :disabled="loading === 'backup'" @click="runBackup">
        手动同步到 GitHub
      </button>
    </section>

    <section class="outputs">
      <article>
        <h2>verify 输出</h2>
        <pre>{{ verifyOutput }}</pre>
      </article>

      <article>
        <h2>stats 输出</h2>
        <pre>{{ statsOutput }}</pre>
      </article>

      <article>
        <h2>backup 输出</h2>
        <pre>{{ backupOutput }}</pre>
      </article>
    </section>
  </main>
</template>
