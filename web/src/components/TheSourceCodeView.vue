<template>
  <div class="border border-slate-300 rounded-md">
    <div class="w-full h-[2rem] rounded-t-md bg-white px-2 border-b border-gray-200">
      <h1 class="my-auto">Source Code</h1>
    </div>
    <table class="w-full table-fixed bg-white text-white font-mono text-sm rounded-b-md">
      <tbody>
        <tr v-for="(line, index) in lines" :key="index" :class="codeLineColor(index)">
          <!-- Line Number -->
          <td class="w-12 text-right pr-4 text-gray-700 border-r border-gray-200">
            {{ index + 1 }}
          </td>
          <!-- Source Code -->
          <td class="pl-4 text-gray-800 whitespace-pre-wrap">
            {{ line }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script>
export default {
  data() {
    return {
      lines: [
        "<?php",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
        "echo 'Hello World!'",
      ],
    }
  },
  methods: {
    codeLineColor(idx) {
      if (idx == 3) {
        return "bg-green-300"
      } else if (idx == 5 || idx == 12) {
        return "bg-yellow-300"
      } else if (idx == 20) {
        return "bg-red-300"
      }
      return ""
    }
  }
}
</script>