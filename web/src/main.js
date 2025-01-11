import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import "./index.css";
import { OhVueIcon, addIcons } from "oh-vue-icons";
import { CoArrowThickFromLeft, CoArrowRight, IoReturnUpBackOutline, PrSearch, HiSolidArrowNarrowDown, FaLongArrowAltDown  } from "oh-vue-icons/icons";

addIcons(CoArrowThickFromLeft, CoArrowRight, IoReturnUpBackOutline, PrSearch, HiSolidArrowNarrowDown, FaLongArrowAltDown);

const app = createApp(App);

app.use(router);
app.component("v-icon", OhVueIcon);
app.mount("#app");
