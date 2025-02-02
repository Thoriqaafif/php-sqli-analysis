<template>
  <div class="h-[100vh] bg-gray-100">
    <header class="h-[10vh] border-b-2">
      <router-link to="/">
        <button class="rounded-lg py-1 px-5 my-auto h-full">
          <v-icon name="io-return-up-back-outline"></v-icon>
          <p class="mx-3 font-mono text-lg font-semibold inline">Back</p>
        </button>
      </router-link>
    </header>
    <main class="h-[90vh] w-screen flex">
      <div class="h-full w-[55vw] px-5 py-3 border-r-2 overflow-y-scroll">
        <h1 class="text-xl font-mono mb-5 font-semibold">Detect {{ data.detected_vulns }} SQLi Vulnerabilities on "{{ data.name }}""</h1>
        <!-- vulnerability cards -->
        <div>
          <div class="flex rounded-md border border-slate-300 w-full h-[7rem] mb-2"
            v-for="(vuln, index) in scanResult.results" :key="index">
            <div class="flex w-[7%] rounded-l-md bg-red-500">
              <h1 class="mx-auto mt-3 text-white font-mono font-semibold text-lg">{{ index + 1 }}</h1>
            </div>
            <div v-if="selectedVuln == index" class="w-full rounded-r-md px-3 py-2 hover:cursor-pointer
             bg-gray-300">
              <h1 class="text-lg font-mono">{{ vuln.path }}</h1>
            </div>
            <div v-else class="w-full rounded-r-md px-3 py-2 hover:cursor-pointer bg-white" @click="selectVuln(index)">
              <h1 class="text-lg font-mono">{{ vuln.path }}</h1>
            </div>
          </div>
        </div>
      </div>
      <div class="h-full w-[45vw] px-2 py-3 overflow-y-scroll">
        <!-- <div class="flex mb-5">
          <div class="border border-slate-800 rounded-lg">
            <button class="px-3 py-1 font-mono rounded-l-lg text-base font-medium" :class="codeBtnClr"
              @click="selectCode">
              Code
            </button>
            <button class="px-3 py-1 font-mono rounded-r-lg text-base font-medium" :class="dataflowBtnClr"
              @click="selectDataflow">
              Dataflow
            </button>
          </div>
        </div> -->
        <div class="flex w-full">
          <!-- code view -->
          <the-source-code-view v-if="viewMode == 'code'"></the-source-code-view>
          <!-- dataflow view -->
          <the-dataflow-view v-else-if="viewMode == 'dataflow'" :trace="scanResult.results[selectedVuln].extra.dataflow_trace"></the-dataflow-view>
        </div>
      </div>
    </main>
  </div>
</template>

<script>
import TheSourceCodeView from '@/components/TheSourceCodeView.vue';
import TheDataflowView from '@/components/TheDataflowView.vue';

export default {
  components: {
    'the-source-code-view': TheSourceCodeView,
    'the-dataflow-view': TheDataflowView,
  },
  computed: {
    codeBtnClr() {
      if (this.viewMode == "code") {
        return "bg-slate-700 text-white"
      }
      return "bg-slate-100 text-black"
    },
    dataflowBtnClr() {
      if (this.viewMode == "code") {
        return "bg-slate-100 text-black"
      }
      return "bg-slate-700 text-white"
    }
  },
  data() {
    return {
      viewMode: "dataflow",
      data: {},
      scanResult: [],
      selectedVuln: 0,
      error: null,
    }
  },
  methods: {
    selectCode() {
      this.viewMode = "code"
    },
    selectDataflow() {
      this.viewMode = "dataflow"
    },
    selectVuln(idx) {
      this.selectedVuln = idx
      console.log(this.selectedVuln)
    },
    async fetchData() {
      // fetch project's vulnerability
      try {
        const id = this.$route.params.id
        const url = `http://localhost:8080/api/project/${id}`
        const response = await fetch(url);
        if (!response.ok) throw new Error('Failed to fetch data');
        const resp = await response.json();
        this.data = resp.data
        this.scanResult = resp.result
        console.log(this.projects)
      } catch (err) {
        this.error = err
        console.log(err)
      }
    }
  },
  beforeMount() {
    this.fetchData()
  }
}
</script>