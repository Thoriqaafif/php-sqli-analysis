<template>
    <div class="h-100vh">
        <header class="h-[25vh]">
            <div
                class="flex justify-between h-[40%] px-5 py-4 border border-x-0 border-t-0 border-b-1 border-slate-400">
                <h1 class="text-xl font-mono font-bold">Your Projects</h1>
            </div>
            <div class="flex h-[60%] justify-between px-5 py-3">
                <div class="relative my-auto w-[40%]">
                    <input v-model="search"
                        class="w-full bg-transparent placeholder:text-slate-400 text-slate-700 text-sm border border-slate-200 rounded-lg pl-3 pr-28 py-2 transition duration-300 ease focus:outline-none focus:border-slate-400 hover:border-slate-300 shadow-sm focus:shadow"
                        placeholder="Search ..." />
                </div>
                <div class="my-auto mx-5">
                    <button class="bg-red-400 rounded-lg py-1 px-5" @click="deleteProjects">
                        <p class="font-mono text-md font-semibold">Delete</p>
                    </button>
                </div>
            </div>
        </header>
        <main class="h-[75vh] p-0 overflow-y-scroll">
            <table class="min-w-full table-auto p-0 m-0">
                <thead class="bg-slate-300 sticky top-0">
                    <tr>
                        <th class="px-4"></th>
                        <th class="w-1/4 px-4  text-left font-mono">Name</th>
                        <th class="px-4 text-left font-mono">Scanned File Counts</th>
                        <th class="px-4 text-left font-mono">Detected Vulnerabilities</th>
                        <th class="px-4 text-left font-mono">Scanned At</th>
                        <th class="px-4 text-left font-mono">Result</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="(project, index) in searchedProjects" :key="project.id" :class="rowClass(index)">
                        <td class="px-4 py-2 top-0">
                            <input type="checkbox" @click="selectProject(index)" :checked="project.selected">
                        </td>
                        <td class="px-4 py-2 text-left font-mono">
                            <router-link :to="'/detail/' + project.id">
                                {{ project.name }}
                            </router-link>
                        </td>
                        <td class="px-4 py-2 font-mono">{{ project.num_of_files }}</td>
                        <td class="px-4 py-2 font-mono">{{ project.detected_vulns }}</td>
                        <td class="px-4 py-2 font-mono">{{ project.created_at.toLocaleString() }}</td>
                        <td class="px-4 py-2 font-mono"><router-link :to="'/detail/' + project.name">
                                <v-icon name="co-arrow-thick-from-left"></v-icon>
                            </router-link></td>
                    </tr>
                </tbody>
            </table>
        </main>
    </div>
</template>

<script>
export default {
    computed: {
        searchedProjects() {
            return this.projects
                .filter((p) => p.name.toLowerCase().startsWith(this.search.toLowerCase()))
        }
    },
    data() {
        return {
            projects: [
                {
                    "id": "2743a42e-b0c1-4b7a-8a6a-32e92262ab4f",
                    "name": "WeBid",
                    "num_of_files": 100,
                    "detected_vulns": 20,
                    "created_at": new Date('August 19, 1975 23:15:30'),
                    "selected": false,
                },
                {
                    "id": "820c707f-755a-4bbd-b595-cd9b103d3e6e",
                    "name": "PHP7-Webchess",
                    "num_of_files": 73,
                    "detected_vulns": 30,
                    "created_at": Date(Date.now()),
                    "selected": false,
                },
                {
                    "id": "160606e9-7c3f-4676-a1de-7d963542a29f",
                    "name": "WeBid",
                    "num_of_files": 100,
                    "detected_vulns": 20,
                    "created_at": Date(Date.now()),
                    "selected": false,
                },
                {
                    "id": "61423e6d-c058-4e34-ab33-9e38b67c7a2e",
                    "name": "PHP7-Webchess",
                    "num_of_files": 73,
                    "detected_vulns": 30,
                    "created_at": Date(Date.now()),
                    "selected": false,
                },
                {
                    "id": "741fa5a2-d79d-45e5-aaee-8c01de81198d",
                    "name": "WeBid",
                    "num_of_files": 100,
                    "detected_vulns": 20,
                    "created_at": Date(Date.now()),
                    "selected": false,
                },
                {
                    "id": "8743674a-b64f-48ca-835c-6151c2cd8337",
                    "name": "PHP7-Webchess",
                    "num_of_files": 73,
                    "detected_vulns": 30,
                    "created_at": Date(Date.now()),
                    "selected": false,
                },
                {
                    "id": "bdd2a287-f460-426e-a828-2cc0fa6b81cd",
                    "name": "WeBid",
                    "num_of_files": 100,
                    "detected_vulns": 20,
                    "created_at": Date(Date.now()),
                    "selected": false,
                },
                {
                    "id": "2fe52757-8eb2-4d29-95f4-4f6604d47845",
                    "name": "PHP7-Webchess",
                    "num_of_files": 73,
                    "detected_vulns": 30,
                    "created_at": Date(Date.now()),
                    "selected": false,
                },
                {
                    "id": "463104d3-8053-4fbf-9737-1fc555249c8a",
                    "name": "WeBid",
                    "num_of_files": 100,
                    "detected_vulns": 20,
                    "created_at": Date(Date.now()),
                    "selected": false,
                },
                {
                    "id": "09018048-1c1c-430a-8e8e-90d8ff6edb7c",
                    "name": "PHP7-Webchess",
                    "num_of_files": 73,
                    "detected_vulns": 30,
                    "created_at": Date(Date.now()),
                    "selected": false,
                }
            ],
            search: "",
        }
    },
    methods: {
        rowClass(idx) {
            if (idx % 2 == 1) {
                return "bg-gray-100"
            }
            return "bg-white"
        },
        selectProject(idx) {
            this.projects[idx].selected = !this.projects[idx].selected
        },
        deleteProjects() {
            this.projects = this.projects.filter((p) => !p.selected)
        }
    },
    beforeMount() {
        try {
            console.log("CheckIcon:", IoCheckboxOutline);
        } catch (error) {
            console.error("Error in beforeMount:", error);
        }
    }
}
</script>