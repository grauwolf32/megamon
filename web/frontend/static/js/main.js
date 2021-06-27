HighlightedReport = Vue.component('h-report', {
    props: ["fragment"],
    render(new_el) {
        var text = this.fragment.text
        var ind  = this.fragment.keywords
        var rootChilds = []

        rootChilds.push(new_el("span", {}, text.substring(0, ind[0])))
        for(var i=0 ;i < ind.length; i++)
        {
            rootChilds.push(new_el("span", {class : "highlight"}, text.substring(ind[i][0], ind[i][1])))
            if(i+1 < ind.length){
                rootChilds.push(new_el("span", {},  text.substring(ind[i][1], ind[i+1][0])))
            }
        }

        rootChilds.push(new_el("span", {},  text.substring(ind[ind.length - 1][1], text.length)))
        return divElement = new_el("div", {class:"text-wrap"}, rootChilds) 
    },
  })

HeadNavigation = Vue.component('h-nav', {
      data: function(){
          return {
            navigation: {
                name : "Megascan",
                pages : [{
                    name: "Github",
                    path: "/github"
                },
                { 
                    name:"Gist",
                    path:"/gist"
                },
                {
                    name:"Settings",
                    path:"/settings"
                },
                {
                    name: "Controls",
                    path: "/controls"
                },]
            }
          }
      },
      template:  `<nav class="navbar navbar-expand-md navbar-dark fixed-top bg-dark">
                  <a class="navbar-brand" href="#">{{navigation.name}}</a>
                  <div class="collapse navbar-collapse" id="navbarCollapse">
                  <ul class="navbar-nav mr-auto">
                     <li v-for="page in navigation.pages" v-bind:class="[page.path == $route.path ? 'nav-item active': 'nav-item']">
                           <router-link class="nav-link" v-bind:to="page.path">{{page.name}}</router-link>
                     </li>
                  </ul></div></nav>`
})

Pagination = Vue.component('p-nav', {
    props: ["pagination"],
    template: `#pnav-template`
})

ModalWindow = Vue.component('v-modal', {
    props: ["content"],
    template: `
    <transition name="modal">
    <div class="modal-mask">
      <div class="modal-wrapper">
        <div class="modal-container">

          <div class="modal-header">
            <h3>Fragment Info</h3>
            <button type="button" class="btn-close btn" aria-label="Close" @click="$emit('close')">x</button>
          </div>

          <div class="modal-body">
            <f-info v-bind:info="content"></f-info>
          </div>

          <div class="modal-footer">
 
          </div>
        </div>
      </div>
    </div>
  </transition>         `
})

RControl = Vue.component('r-control',{
    props:["fragment"],
    data : function(){
        return{
            buttons : [
                {name: "Verify", action: 2},
                {name: "Close",  action: 1},
                {name: "Info",   action: 0}
            ]
        }
    },
    template: `<div class="btn-group">
                    <button v-for="button in buttons" v-on:click="$emit('markResult', [button.action, fragment.id])"  type="button" class="btn btn-outline-primary">{{button.name}}</button>
               </div>`
})

VItems = Vue.component('v-items',{
    props:["vitem"],
    data: function(){
        return{
            selection: [],
            copy:"",
        }
    },
    methods:{
        remove: function(){
            this.$emit("remove",{id: this.vitem.id, selected:this.copy})
        },
        add: function(){
            this.$emit("add",{id: this.vitem.id, selected:this.copy})
        },
        copyval: function(){
            this.copy = this.selection[0]
            this.$emit("select", {id: this.vitem.id, selected:this.copy})
        }
    },
    template: `<div>
                    <label class="form-label"> {{vitem.name}}</label><br>
                    <div class="input-group mb-3">
                        <select class="form-select select-item" multiple v-model:value="selection" v-on:change="copyval()">
                            <option class="list-group-item" v-for="item in vitem.data"> {{ item }} </option>
                        </select>
                    </div>

                <div class="input-group mb-3">
                <input type="text" class="form-control input-item" v-model:value="copy"></input>
                <div class="btn-group">
                    <button type="button" class="btn btn-outline-primary" v-on:click="add()">Add</button>
                    <button type="button" class="btn btn-outline-primary" v-on:click="remove()">Remove</button>
                </div>
                </div>
            </div>`
})

FragmentInfo = Vue.component('f-info', {
    props : {
        info :{
            name: {
                type: String
            },
            path:{
                type: String
            },
            html_url: {
                type: String
            },
            repository:{
                name : {
                    type: String
                },
                owner:{
                    login:{
                        type: String
                    },
                    url:{
                        type: String
                    }
                }
            }
        }
    },
    template: `
    <div>
    <ul>
        <li><b>Filename</b>: {{info.name}}</li>
        <li><b>Full path</b>: {{info.path}}</li>
        <li><b>Link: </b><a v-bind:href="info.html_url" target="_blank">{{info.name}}</a></li>
        <li><b>Repository:</b> {{info.repository.name}}</li>
        <li><b>Owner:</b> <a v-bind:href="info.repository.owner.url" target="_blank"> {{info.repository.owner.login}}</a></li>
    </ul>
    </div>
    `
})

Controls = Vue.component('controls',{
    data : function(){
        return{
            statuses: {"github":"unknown"},
            polling : ''
        }
    },
    methods:{
        getTasksAvailable: function(){
            var requestURI = "/leaks/api/task/available"
            axios.get(requestURI)
                .then(response => {
                    if(response.status == 200){
                        for(var i in response.data){
                            element = response.data[i]
                            this.statuses[element] = "unknown"
                        }
                    }
                    console.log(this.statuses)
                    this.updateStatuses()
                })
                .catch(error => {
                    console.log(error)
                })
        },
        updateStatuses: function(){
            var requestURI = ""
            for(var element in this.statuses){
                requestURI = "/leaks/api/task/"+element+"/info"
                axios.get(requestURI)
                    .then(response => {
                        if(response.status == 200){
                            console.log("element: " + element)
                            this.statuses[element] = response.data.toString()
                            console.log("value: "+this.statuses[element])
                        }
                    })
                    .catch(error => {
                        console.log(error)
                    })
            }
        },
        taskManager: function(data){
            var task = data[0]
            var status = data[1]

            console.log(data)

            var requestURI = "/leaks/api/task/"+task+"/"+status
            axios.get(requestURI)
                .then(response => {
                    if(status == "info"){
                        this.statuses[task] = response.data
                    } else {
                        requestURI = "/leaks/api/task/"+task+"/info"
                        axios.get(requestURI)
                            .then(response => {
                                if(response.status == 200){
                                    this.statuses[task] = response.data
                                }
                            })
                            .catch(error => {
                                console.log(error)
                            })
                    }
                })
                .catch(error => {
                    console.log(error)
                })
        },
    },
    created: function() {
        this.getTasksAvailable()
        this.polling = setInterval(this.updateStatuses(), 15000)
    },
    beforeDestroy: function(){
        clearInterval(this.polling)
    },
    template: `
    <div>
    <br/><br/><br/><br/>
    <table>
        <tr>
            <td v-for="(status, task) in statuses" v-bind:key="task">
                <h3>{{task}} ({{ status }})</h3>
                <div class="btn-group">
                    <button type="button" class="btn btn-outline-primary" v-on:click="taskManager([task, 'start'])"> start </button>
                    <button type="button" class="btn btn-outline-primary" v-on:click="taskManager([task, 'stop'])"> stop </button>
                </div>
            </td>
        </tr>
    </table>
    </div>
    `
})


Settings = Vue.component('settings', {
    data : function(){
        return {
            settings: {
                db_credentials : {name: "", database: "", password: ""},
                github : {tokens :[],  langs : { blacklist: []}},
                globals : {keywords : {}, rules:{}}
            },
            selected: "",
            checkbox: false,
            rules:[],
            keywords:[]
        }
    },
    methods:{
        getSettings: function(){
            var requestURI = '/leaks/api/settings'
            axios.get(requestURI)
                .then(response => {
                    this.settings = response.data
                    for(var keyword in this.settings.globals.keywords){
                        this.keywords.push(keyword)
                    }

                    for(var rule in this.settings.globals.rules){
                        this.rules.push(rule)
                    }

                    console.log(this.settings)
                    console.log(this.rules)
                    console.log(this.keywords)

                })
                .catch(error => {
                    console.log(error)
                })
        },
        add: function(data){
            var itemId = data.id
            var selected = data.selected

            if(itemId == 1){
                this.settings.github.tokens.push(selected)
            } else if (itemId == 2){
                this.settings.github.langs.blacklist.push(selected)
            } else if (itemId == 3){
                var type = 0 // searchable keyword
                if (!this.checkbox){
                    type = 1 // inner keyword
                }

                this.keywords.push(selected)
                this.settings.globals.keywords[selected] = {
                    "value" : selected,
                    "id": 0,
                    "type":type
                }
            } else if (itemId == 4){
                this.rules.push(selected)
                this.settings.globals.rules[selected]= {
                    "rule" : selected,
                    "name" : "regexp",
                    "id":0,
                }
            }
        },
        remove: function(data){
            var itemId = data.id
            var selected = data.selected

            if(itemId == 1){
                elId = this.settings.github.tokens.indexOf(selected)
                if(elId != -1){
                    this.settings.github.tokens.splice(elId, 1)
                }
            } else if (itemId == 2){
                elId = this.settings.github.langs.blacklist.indexOf(selected)
                if(elId != -1){
                    this.settings.github.langs.blacklist.splice(elId, 1)
                }
            } else if (itemId == 3){
                elId = this.keywords.indexOf(selected)
                
                if(elId != -1){
                    this.keywords.splice(elId, 1)
                    delete this.settings.globals.keywords[selected]
                }
            } else if (itemId == 4){
                elId = this.rules.indexOf(selected)
                if(elId != -1){
                    this.rules.splice(elId, 1)
                    delete this.settings.globals.rules[selected]
                }
            }
        },
        select: function(data){
            var itemID = data.id
            this.selected = data.selected

            console.log(data)
            
            if(itemID == 3){
                for(keyword in this.settings.globals.keywords){
                    if(keyword == this.selected){
                        var type = this.settings.globals.keywords[keyword].type
                        if (type == 0){
                            this.checkbox = true
                        } else{
                            this.checkbox = false
                        }
                    }
                }
            }
        },
        update: function(){
            console.log(this.settings)
            var requestURI = "/leaks/api/settings"
            axios.post(requestURI, this.settings)
        },
    },
    created : function(){
        this.getSettings()
    },
    template: "#settings-template"
})


Fragments = Vue.component('r-fragments', {
    props: ["pagetype"],

    data: function() {
        return {
            fragments : [],
            pagination: {
                pages:[],
                maxPage: 0,
                currentPage : 0
            },
            modal : {
                content: "",
                show : false
            },
            reportStatuses:[
                {name: "New",    value: "0"},
                {name: "Closed", value: "1"},
                {name: "Verified",    value: "2"},
                {name: "Autoremoved", value: "3"},
            ],
            availableLimits:[10, 20, 50, 100],
            reportStatus: "0",
            limit: 10
        } 
    },
    
    template: "#fragments-template",
    methods: {
            updatePage: function () {
            offset = this.pagination.currentPage*this.limit
           
            //Get fragments
            var requestURI = '/leaks/api/report/frags/' + this.pagetype + "/" +  this.reportStatus + '?limit=' + this.limit + '&offset=' + offset
            axios.get(requestURI)
                .then(response => {
                    this.fragments = response.data
                })
                .catch(error => {
                    console.log(error)
                })

            //Get fragments count
            requestURI = '/leaks/api/report/count/' + this.pagetype + "/" +  this.reportStatus + '?limit=' + this.limit + '&offset=' + offset
            axios.get(requestURI)
                .then(response => {
                    var nResults = response.data["count"]
                    this.updatePagination(nResults)
                })
                .catch(error => {
                    console.log(error)
                })
            },
            showInfo(fragmentId){
                var fragment = {}
                for(var i=0;i < this.fragments.length;i++){
                    if(this.fragments[i].id == fragmentId){
                        fragment = this.fragments[i]
                        break
                    }
                }
                
                var requestURI = '/leaks/api/report/info/' + fragmentId
                
                axios.get(requestURI)
                    .then(response => {
                        this.modal.content = response.data
                    })
                    .catch(error => {
                        console.log(error)
                    })

                this.modal.show = true
            },
        
        markResult: function(data){
            var fragment_id = data[1]
            var status = data[0]
            var requestURI = "/leaks/api/report/mark/"+ fragment_id + "/"+data[0]

            console.log(data)
            if(status == 0){
                this.showInfo(fragment_id)
                return
            } 
            var fid = -1
            var rid = -1

            for(var i = 0 ;i < this.fragments.length; i++){
                if (this.fragments[i].id == fragment_id){
                    fid = i
                    rid = this.fragments[i].report_id
                    break
                }
            }

            console.log("fid: " + fid +" rid: "+rid)

            if(status == 1){
                axios.get(requestURI).then(respons=>{
                    if(respons.status == 200 && fid != -1){
                        this.fragments.splice(fid, 1)
                        if(this.fragments.length == 0){
                            this.updatePage()
                        }
                        return
                    }
                })
            } else if(status == 2 && rid != -1){
                axios.get(requestURI).then(
                    response => {
                        if(response.status == 200){
                            var splids = []
                            for(var i = 0 ;i < this.fragments.length; i++){
                                if(this.fragments[i].report_id == rid){
                                    splids.push(i)
                                }
                            }
                            var newFragments = []
                            for(var i = 0 ;i < this.fragments.length; i++){
                                if(splids.indexOf(i) == -1){
                                    newFragments.push(this.fragments[i])
                                }
                            }
                            this.fragments = newFragments
                        }
                    }
                )
            }
        },

        updatePagination: function(nResults){
            this.pagination.maxPage =  Math.ceil(nResults / this.limit) 
            currentPage = this.pagination.currentPage
            maxPage = this.pagination.maxPage

            var st  = 0
            var end = 0

            if (maxPage - currentPage >= 5 && currentPage >= 5){
                st  = currentPage - 5
                end = currentPage + 5
            } else if(currentPage < 5){
                var nLeft = 5 - currentPage
                if (maxPage - currentPage >= 5 + nLeft){
                    end = currentPage + 5 + nLeft
                } else {
                    end = maxPage
                }
            } else {
                st = currentPage - 5
                end = maxPage
            }

            var pagination = []
            for(var i=st; i < end; i++){
                pagination.push({id: i})
            }

            this.pagination.pages = pagination
            return
        },
        goTo: function(currentPage){
            if (currentPage > this.pagination.maxPage){
                this.pagination.currentPage = this.maxPage
            }else if (currentPage < 0){
                this.pagination.currentPage = 0
            }else {
                this.pagination.currentPage = currentPage
            }
            this.updatePage()
        },

        skipRight: function(){
            this.goTo(this.pagination.currentPage+10)
        },
        skipLeft: function(){
            this.goTo(this.pagination.currentPage-10)
        }
    },
    watch: {
        $route : function(){
            this.updatePage()
        }
    },
    created: function(){
        this.updatePage()
        return
    },
})


const router = new VueRouter({
    routes :[ 
        {path: "/", component:Fragments, props:{pagetype:"github"}},
        {path: "/github", component:Fragments, props:{pagetype:"github"}},
        {path: "/gist",  component:Fragments, props:{pagetype:"gist"}},
        {path: "/settings",  component:Settings },
        {path: "/controls", component:Controls },
    ],
    mode: "history"
})

var app = new Vue({
    el: '#app',
    data: {
        pagetype: "github",
    },
    components:{
        'report-highlight': HighlightedReport,
        'header-navigation': HeadNavigation,
        'pagination-navigation' : Pagination,
        'report-control' : RControl,
        'fragments' : Fragments,
        'settings' : Settings,
        'controls' : Controls,
        'v-items' : VItems,
        'v-modal':ModalWindow,
        'f-info':FragmentInfo,
    },
    router: router
})