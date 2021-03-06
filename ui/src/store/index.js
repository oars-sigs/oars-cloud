import Vue from 'vue'
import Vuex from 'vuex'


Vue.use(Vuex)

export const store = new Vuex.Store({
  state: {
    namespaces: [],
    currentNamespace: "",
  },
  getters: {
    getCurrentNamespace: state => state.currentNamespace,
  },
  mutations: {
    NamepaceList(state,{params}){
        state.namespaces=params;
    },
    SetCurrentNamespace(state,namespace){
        state.currentNamespace=namespace
    },
  },
  actions: {
    
  }
})