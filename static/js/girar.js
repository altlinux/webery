angular.module('girar', ['ngRoute', 'ngSanitize','ui.bootstrap','ui.chart','ui.bootstrap.typeahead','infinite-scroll'])
.config(['$routeProvider', '$locationProvider',
	function($routeProvider, $locationProvider) {
		$routeProvider
			.when('/task/:taskId', {
				templateUrl: '/task.html',
			})
			.when('/taskpkgs/:repo/:state/:taskId', {
				templateUrl: '/task-packages.html',
			})
			.when('/acl/:repo/:type/:name', {
				templateUrl: '/acl-show.html',
			})
			.when('/acl/:repo/:type', {
				templateUrl: '/acl-search.html',
			})
			.when('/acl-nobody/:repo', {
				templateUrl: '/acl-packages-nobody.html',
			})
			.when('/dashboard', {
				templateUrl: '/dashboard.html',
			})
			.when('/acl', {
				templateUrl: '/acl-search.html',
			})
			.when('/main', {
				templateUrl: '/main.html',
			})
			.when('/log/:url*', {
				templateUrl: '/log.html'
			})
			.when('/apidoc', {
				templateUrl: '/apidoc.html'
			})
			.otherwise({
				redirectTo: '/main'
			});
		$locationProvider.html5Mode(true);
	}
])
.filter('convertState', function ($sce) {
	return function(state) {
		var label = "default";
		switch (state) {
			case "eperm":
			case "failed":
				label = "danger";
				break;
			case "tested":
			case "done":
				label = "success";
				break;
			case "new":
			case "awaiting":
			case "postponed":
			case "building":
			case "pending":
			case "commiting":
				label = "info"
				break;
			case "delete":
				break;
			default:
				return "";
		}
		var out = '<span class="label label-' + label + '">' + state + '</span>';
		return $sce.trustAsHtml(out);
	};
})
.filter('convertBool', function ($sce) {
	return function (value,arg) {
		var out = "";
		if (value === true) {
			out = '<span class="label label-default">'+arg+'</span>';
		}
		return $sce.trustAsHtml(out);
	};
})
.filter('firstLetter', function () {
	return function (value) {
		if (value != null) {
			return value.charAt(0);
		}
		return '';
	};
})
.factory('taskState', ['$http', function($http) {
	var tasks = {
		new:       [],
		awaiting:  [],
		building:  [],
		pending:   [],
		commiting: []
	};

	list = function(val, limit) {
		return $http.get('/api/v1/tasks', {
			params: {
				state: val,
				limit: limit || 10
			}
		}).then(function(response) {
			if (!response.data.data) {
				return [];
			}

			tasks[val] = response.data.data.result.map(function(item) {
				item.url = "/task/" + item.taskid;
				item.include = "list-task-" + val + ".html";
				return item;
			}).sort(function (a, b) {
				if (a.taskid < b.taskid) {
					return 1;
				}
				if (a.taskid > b.taskid) {
					return -1;
				}
				return 0;
			});
		});
	};
	
	get = function() {
		return tasks;
	};

	return {
		  get:  get,
		  list: list
	};
}])
.controller('BodyCtrl', ['$scope', '$rootScope', function($scope, $rootScope) {
	$rootScope.GitAltUrl = "//git.altlinux.org";
}])
.controller('MainCtrl', ['$route', '$routeParams', '$location', function($route, $routeParams, $location) {
	this.$route       = $route;
	this.$location    = $location;
	this.$routeParams = $routeParams;
}])
.controller('ApiDocCtrl', ['$rootScope', function($rootScope) {
	$rootScope.setActive('apidoc');
}])
.controller('MenuCtrl', ['$rootScope', function($scope) {
	var active = '';
	$scope.isActive  = function(val) { return active === val; };
	$scope.setActive = function(val) { active = val; };
}])
.controller('LogCtrl', ['$scope','$routeParams', '$http', '$location', '$anchorScroll', function($scope, $routeParams, $http, $location, $anchorScroll) {
	$scope.url = $routeParams.url || "";
	$scope.log = [];
	$scope.skipped = false;
	$scope.showend = false;

	linenumber = 0;
	arr = [];
	arrlimit = 100;

	isError = function(s) {
		if (!s) {
			return false;
		}
		if (s.search(/ :: acl check FAILED/) != -1) {
			return true;
		}
		if (s.search(/ :: gears inheritance check FAILED/) != -1) {
			return true;
		}
		if (s.search(/ :: dependencies check FAILED/) != -1) {
			return true;
		}
		if (s.search(/^find-requires: ERROR/) != -1) {
			return true;
		}
		if (s.search(/^make\[[0-9]+\]: .* Error 1/) != -1) {
			return true;
		}
		if (s.search(/^configure: error: /) != -1) {
			return true;
		}
		return false;
	}

	$scope.gotoTop = function() {
		window.scrollTo(0, 0);
	};

	$scope.showBegin = function() {
		$scope.skipped = false;
		$scope.log = [];
		$scope.appendMore();
	};

	$scope.showEnd = function() {
		var i = arr.length - arrlimit;
		if (i < 0) {
			return;
		}
		$scope.log = [{
			index:   i,
			number:  i+1,
			select:  (linenumber === i+1),
			error:   isError(arr[i]),
			content: arr[i]
		}];
		$scope.skipped = true;
		$scope.appendMore();
	};

	$scope.prependMore = function() {
		if ($scope.log.length > 0 && $scope.log[0].index === 0) {
			return;
		}
		var limit = 10;
		var i = $scope.log[0].index;
		for (i--; i >= 0 && limit > 0; i--) {
			$scope.log.unshift({
				index:   i,
				number:  i+1,
				select:  (linenumber === i+1),
				error:   isError(arr[i]),
				content: arr[i]
			});
			limit--;
		}
		$scope.skipped = (i >= 0);
	};

	$scope.appendMore = function() {
		var limit = arrlimit;
		var i = 0;

		if ($scope.log.length > 0) {
			i = $scope.log[$scope.log.length - 1].index + 1;
		}

		for (; i < arr.length && limit > 0; i++) {
			$scope.log.push({
				index:   i,
				number:  i+1,
				select:  (linenumber === i+1),
				error:   isError(arr[i]),
				content: arr[i]
			});
			limit--;
		}
	};

	$http.get('/rawlog/'+$scope.url, {
		params: {},
	}).then(function(response) {
		arr = response.data.split('\n');

		if (arr.length > arrlimit) {
			$scope.showend = true;
		}

		var limit = arrlimit;
		var pos = 0;

		if (location.hash === "") {
			for (var i = 0; i < arr.length; i++) {
				if (isError(arr[i]) === false) {
					continue;
				}
				if (arrlimit < i) {
					$scope.skipped = true;
					pos = i;
				}
				break;
			}
		} else {
			if (location.hash.startsWith("#L")) {
				linenumber = parseInt(location.hash.replace(/[^0-9]/g, ''), 10);
				$scope.skipped = true;
				pos = linenumber - 1;
			}
			if (location.hash.startsWith("#end")) {
				$scope.skipped = true;
				pos = arr.length - arrlimit;
			}
		}

		for (; pos < arr.length && limit > 0; pos++) {
			$scope.log.push({
				index:   pos,
				number:  pos+1,
				select:  (linenumber === pos+1),
				error:   isError(arr[pos]),
				content: arr[pos]
			});
			limit--;
		}

	},
	function(reason) {
		alert("Error: " + reason.statusText);
		window.location = document.referrer;
	});
}])
.controller('SearchCtrl', ['$scope','$rootScope', '$location', '$http', function($scope, $rootScope, $location, $http) {
	$rootScope.setActive('main');

	$scope.getResults = function(val) {
		return $http.get('/api/v1/search', {
			params: {
				prefix: val,
				limit: 10,
			}
		}).then(function(response) {
			if (!response.data.data) {
				return [];
			}

			var fields = ["taskid", "owner", "pkgname", "repo"];
			var re = new RegExp('^(' + response.data.data.query[0]["search.key"]["$regex"] + ')');

			return response.data.data.result.map(function(item) {
				item.task_id = item.taskid;
				item.url = "/task/" + item.taskid;
				item.include = "search-suggestion-" + item.objtype + ".html";

				fields.forEach(function (element, index, array) {
					if (item[element]) {
						item[element] = item[element].toString().replace(re, '<span class="searchmatch">$1</span>');
					}
				});

				return item;
			}).sort(function (a, b) {
				if (a.taskid < b.taskid) {
					return 1;
				}
				if (a.taskid > b.taskid) {
					return -1;
				}
				return 0;
			});
		});
	};
	$scope.showResult = function(item, model, label) {
		$location.url(item.url);
	};
}])
.controller('TaskCtrl', ['$routeParams', '$scope', '$rootScope', '$http', function($routeParams, $scope, $rootScope, $http) {
	this.name = "TaskCtrl";
	this.params = $routeParams;

	getTask = function() {
		var taskid = $routeParams.taskId;

		return $http.get('/api/v1/tasks/' + taskid, {}).then(function(response) {
			$scope.task     = response.data.data.result;
			$scope.task.tries = [];
			for (var i = $scope.task.try; i; i--) {
				$scope.task.tries.push(i);
			}

			$scope.subtasks = [];
			$http.get('/api/v1/tasks/' + taskid + "/subtasks", {}).then(function(response) {
				var subtasks = response.data.data.result;
				subtasks.map(function(item) {
					if (item.status === 'cancelled') {
						return;
					}
					if (item.pkgname === '' && item.dir !== '') {
						var pos = item.dir.lastIndexOf("/");
						if (pos !== -1) {
							item.pkgname = item.dir.substring(pos+1);
						}
					}
					item.active = (item.status === 'active');
					item.include = "subtask-" + item.type + ".html";
					$scope.subtasks.push(item);
				});
			});
		});
	};
	getTask();
}])
.controller('AclInfoCtrl', ['$routeParams', '$scope', '$rootScope', '$http', function($routeParams, $scope, $rootScope, $http) {
	if ($routeParams.type != "groups" && $routeParams.type != "packages") {
		alert("Wrong type: " + $routeParams.type);
		return;
	}

	var repo = $routeParams.repo;
	var type = $routeParams.type;
	var name = $routeParams.name;

	$scope.Name     = "";
	$scope.Repo     = "";
	$scope.Members  = [];
	$scope.Found    = false;
	$scope.NotFound = false;

	getACL = function() {
		return $http.get('/api/v1/acl/' + repo + '/' + type + '/' + name, {
			params: {}
		}).then(
			function(response) {
				$scope.Found   = true;
				$scope.Name    = $routeParams.name;
				$scope.Repo    = $routeParams.repo;
				$scope.Members = response.data.data.result.members;
				$scope.Members.map(function(item) {
					item.include = "acl-show-" + item.type + ".html";
				});
				$scope.Members.sort(function (a, b) {
					if (a.leader) {
						return -1;
					}
					return (
						b.type.localeCompare(a.type) ||
						a.name.localeCompare(b.name)
					);
				});
			},
			function (response) {
				$scope.NotFound = true;
				$scope.Name     = $routeParams.name;
				$scope.Repo     = $routeParams.repo;
			}
		);
	};

	getACL();

	if ($routeParams.type === "groups") {
		if ($routeParams.name === "everybody" || $routeParams.name === "nobody") {
			$scope.MemberInclude = "acl-show-group-" + $routeParams.name + ".html";
		}
	}
}])
.controller('AclSearchCtrl', ['$routeParams', '$scope', '$rootScope', '$location', '$http', function($routeParams, $scope, $rootScope, $location, $http) {
	$rootScope.setActive('acl');

	$scope.Type = $routeParams.type || "packages";
	$scope.Repo = $routeParams.repo || "sisyphus";
	$scope.Prefix = "";
	$scope.Repos = [];
	$scope.Packages = [];

	$scope.toggleRepo = function(repo) {
		$scope.Repo = repo;

		if ($scope.Prefix.length != 0) {
			$scope.getResults($scope.Prefix);
		}
	};

	$scope.toggleType = function(type) {
		$scope.Type = type;

		if ($scope.Prefix.length != 0) {
			$scope.getResults($scope.Prefix);
		}
	};

	$scope.getResults = function(val) {
		$scope.Prefix = val.toLowerCase();

		return $http.get('/api/v1/acl/' + $scope.Repo + '/' + $scope.Type, {
			params: {
				prefix: $scope.Prefix,
				limit: 10
			}
		}).then(function(response) {
			if (!response.data.data) {
				return [];
			}

			var re  = new RegExp('^(' + response.data.data.query[0]["name"]["$regex"] + ')', 'i');

			return response.data.data.result.map(function(item) {
				item.url = "/acl/" + $scope.Repo + "/" + $scope.Type + "/" + item.name;
				item.namematch = item.name.replace(re, '<span class="searchmatch">$1</span>');
				item.members.map(function(item) {
					item.include = "acl-show-" + item.type + ".html";
				});
				return item;
			}).sort(function (a, b) {
				return a.name.localeCompare(b.name);
			});
		});
	};

	$scope.showResult = function(item, model, label) {
		$location.url(item.url);
	};

	getRepos = function() {
		return $http.get('/api/v1/acl/', {
			params: {}
		}).then(function(response) {
			$scope.Repos = response.data.data;
		});
	};

	getRepos();
}])
.controller('AclNobodyCtrl', ['$routeParams', '$scope', '$rootScope', '$http', function($routeParams, $scope, $rootScope, $http) {
	$scope.Repo = $routeParams.repo || "sisyphus";
	$scope.Num      = 0;
	$scope.Alphabet = {};
	$scope.Found    = false;
	$scope.NotFound = false;

	getPackages = function() {
		return $http.get('/api/v1/acl/' + $scope.Repo + '/packages', {
			params: {
				member: "nobody"
			}
		}).then(
			function(response) {
				if (!response.data || !response.data.data || response.data.data.result.length == 0) {
					$scope.NotFound = true;
					return;
				}
				$scope.Num = response.data.data.result.length;
				$scope.Found = true;
				response.data.data.result.map(function(item) {
					var ch = item.name.charAt(0);
					if (!$scope.Alphabet.hasOwnProperty(ch)) {
						$scope.Alphabet[ch] = {};
						$scope.Alphabet[ch]['char'] = ch;
						$scope.Alphabet[ch]['packages'] = [];
					}
					$scope.Alphabet[ch]['packages'].push(item.name);
				});
			},
			function (response) {
				$scope.NotFound = true;
			}
		);
	};

	getPackages();
}])
.controller('TackPkgsCtrl', ['$scope', '$rootScope', '$routeParams', '$http', function($scope, $rootScope, $routeParams, $http) {
	$scope.repo = $routeParams.repo;
	$scope.taskid = $routeParams.taskId;
	$scope.state = $routeParams.state;
	$scope.pkgs = {};

	pkg = function(mode, name, evr) {
		if (!$scope.pkgs[name]) {
			$scope.pkgs[name] = {
				Name: name
			};
		}
		if (!$scope.pkgs[name].EVRs) {
			$scope.pkgs[name].EVRs = {};
		}
		$scope.pkgs[name].EVRs[mode] = evr;
	};

	$http.get('/rawlog/'+$scope.repo+'/'+$scope.state+'/'+$scope.taskid+'/plan/add-bin', {
		params: {},
	}).then(function(response) {
		var arr = response.data.split('\n');
		for (var i = 0; i < arr.length; i++) {
			var fields = arr[i].split('\t');
			if (fields.length === 6) {
				pkg("added", fields[0], fields[1]);
			}
		}
	},
	function(reason) {
		alert("Error: " + reason.statusText);
	});

	$http.get('/rawlog/'+$scope.repo+'/'+$scope.state+'/'+$scope.taskid+'/plan/rm-bin', {
		params: {},
	}).then(function(response) {
		var arr = response.data.split('\n');
		for (var i = 0; i < arr.length; i++) {
			var fields = arr[i].split('\t');
			if (fields.length === 4) {
				pkg("removed", fields[0], fields[1]);
			}
		}
	});
}])
.controller('DashBoardCtrl', ['$scope', '$rootScope', '$http', 'taskState', '$timeout', function($scope, $rootScope, $http, $taskState, $timeout) {
	$rootScope.setActive('dashboard');

	var defRepo = 'sisyphus';
	var taskLimit = 1000000;
	var updatePeriod = 60000;
	var series  = ['awaiting', 'building', 'pending', 'committing'];

	$scope.taskTemplate = "dashboard-task.html";
	$scope.lastUpdate = 0;
	$scope.data = [[0]];
	$scope.chartOptions = {};

	$scope.refresh = {
		_graph:     false,
		awaiting:   false,
		building:   false,
		pending:    false,
		committing: false
	};

	getQueue = function() {
		return $http.get('/api/unversioned/statistic/queue', {
			params: {}
		}).then(function(response) {
			var res = response.data.data.result;

			var width = $('#chart').width();
			var rendererOptions = {};
			var repos = [];
			
			Object.keys(res).sort(function(a, b) {
				if (a === defRepo) {
					return -1;
				}
				return b.localeCompare(a);
			}).forEach(function(repoE) {
				if (repos.length >= 1) {
					var sum = 0;

					series.forEach(function(seriesE) {
						sum += res[repoE][seriesE];
					});

					if (sum === 0) {
						return;
					}
				}
				repos.push(repoE);
			});

			if (repos.length < 4) {
				rendererOptions.barWidth = Math.round(((width / 3) * 20) / 100);
			}

			var plotSeries = [];
			$scope.data = [];

			series.forEach(function(seriesE, seriesI) {
				$scope.data[seriesI] = [];
				plotSeries[seriesI] = {label: seriesE};
			});

			repos.forEach(function(repoE, repoI) {
				series.forEach(function(seriesE, seriesI) {
					  $scope.data[seriesI][repoI] = res[repoE][seriesE];
				});
			});

			$scope.chartOptions = { 
				seriesDefaults: {
					renderer: $.jqplot.BarRenderer,
					rendererOptions: rendererOptions,
					pointLabels: {
						show: true,
						formatString: '%d'
					}
				},
				series: plotSeries,
				legend: {
					show: true,
					location: 'ne',
					rowSpacing: '10px',
					placement: 'outside'
				},
				axes: {
					xaxis: {
						renderer: $.jqplot.CategoryAxisRenderer,
						ticks: repos
					},
					yaxis: {
						tickInterval: 1
					}
				}
		    };
		});
	};

	refreshGraph = function() {
		$scope.refresh['_graph'] = true;
		$scope.lastUpdate = new Date();
		$scope.tasks = $taskState.get();

		getQueue().then(function() {
			$scope.refresh['_graph'] = false;
		});

		$timeout(refreshGraph, updatePeriod)
	};

	refreshTasks = function(state) {
		$scope.refresh[state] = true;
		$taskState.list(state, taskLimit).then(function() {
			$scope.refresh[state] = false;
		});
		$timeout(function() { refreshTasks(state) }, updatePeriod);
	};

	refreshGraph();

	for (var i = 0; i < series.length; i++) {
		refreshTasks(series[i]);
	}
}])
;
