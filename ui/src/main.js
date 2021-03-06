import 'core-js/stable';
import '@mdi/font/css/materialdesignicons.css';
import Vue from 'vue';
import Vuetify from 'vuetify';
import 'vuetify/dist/vuetify.min.css';
import Message from './components/Message/index.js'
import { fetch } from "@/utils/request";
import App from './App.vue';
import router from './router/router'
import {store} from './store'

Vue.prototype.$message = Message
Vue.prototype.$call = fetch
Vue.use(Vuetify);

Vue.config.productionTip = false;
Vue.config.devtools = true;

router.beforeEach((to, from, next) => {
	if (to.meta.title) {
		document.title = to.meta.title
	}
	next()
})


new Vue({
	router,
	store,
	vuetify: new Vuetify({
		icons: {
			iconfont: 'mdi'
		},
		theme: {
			dark: false
		}
	}),
	render: h => h(App),
	mounted() {
		//this.$router.replace('/')
	},
}).$mount('#app');