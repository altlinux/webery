angular.module('girar', ['ngRoute', 'ngSanitize','relativeDate','ui.bootstrap','ui.bootstrap.typeahead'])
.config(['$routeProvider', '$locationProvider',
	function($routeProvider, $locationProvider) {
		$routeProvider
			.when('/task/:taskId', {
				templateUrl: '/task.html',
				controller: 'TaskCtrl',
//				controllerAs: 'task'
			})
			.when('/acl/:repo/:type/:name', {
				templateUrl: '/acl-show.html',
				controller: 'AclInfoCtrl',
			})
			.when('/acl/:repo/:type', {
				templateUrl: '/acl-packages.html',
				controller: 'AclInfoCtrl',
			})
			.when('/acl-nobody/:repo', {
				templateUrl: '/acl-packages-nobody.html',
				controller: 'AclNobodyCtrl',
			})
			.when('/acl', {
				templateUrl: '/acl-packages.html',
				controller: 'AclCtrl',
			})
			.when('/suggestion', {
				controller: 'SuggestionCtrl',
			})
			.when('/main', {
				templateUrl: '/main.html',
				controller: 'SearchCtrl',
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
			default:
				return "";
		}
		var out = '<span class="label label-' + label + '">' + state + '</span>';
		return $sce.trustAsHtml(out)
	};
})
.filter('convertShared', function ($sce) {
	return function (value) {
		var out = "";
		if (value === true) {
			out = '<span class="label label-default">shared</span>';
		}
		return $sce.trustAsHtml(out)
	};
})
.filter('convertTestonly', function ($sce) {
	return function (value) {
		var out = "";
		if (value === true) {
			out = '<span class="label label-default">test only</span>';
		}
		return $sce.trustAsHtml(out)
	};
})
.directive('focusItem', function($timeout) {
	return {
		link: function(scope, element, attrs) {
			scope.$watch(attrs.focusItem, function() {
				element[0].focus();
			});
		}
	};
})
.controller('BodyCtrl', ['$scope', '$rootScope', function($scope, $rootScope) {
	$rootScope.GitAltUrl = "//git.altlinux.org";
}])
.controller('MainCtrl', ['$route', '$routeParams', '$location', function($route, $routeParams, $location) {
	this.$route       = $route;
	this.$location    = $location;
	this.$routeParams = $routeParams;
}])
.controller('ApiDocCtrl', ['$scope', function($scope) {
	$scope.oneAtATime = true;
	$scope.status = {};

	$scope.toggle_open = function(i) {
		var name = "group-" + i;
		for (var prop in $scope.status) {
			if (!$scope.status.hasOwnProperty(prop)) {
				continue;
			}
			if (prop === name) {
				continue;
			}
			$scope.status[prop] = false;
		}
		var old = $scope.status[name] || false;
		$scope.status[name] = !old;
	};

	$scope.is_open = function(i) {
		var name = "group-" + i;
		return $scope.status[name] || false;
	};

}])
.controller('SearchCtrl', ['$scope', '$location', '$http', function($scope, $location, $http) {
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

	getTask = function(taskid) {
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
					item.active = (item.status === 'active');
					switch (item.type) {
						case "delete":
						case "copy":
							item.Built = false;
							item.SourceURL = "" +
								$rootScope.GitAltUrl + "/gears/" +
								item.pkgname.charAt(0) + "/..git?p=" +
								item.pkgname + ".git;a=shortlog;h=refs/heads/" +
								$scope.task.repo;
							break;
						default:
							item.Built = true;
							item.SourceURL = "" +
								$rootScope.GitAltUrl + "/tasks/" +
								item.taskid + "/gears/" +
								item.subtaskid + "/git";
					}
					$scope.subtasks.push(item)
				});
			});
		});
	};

	getTask($routeParams.taskId);
}])
.controller('AclCtrl', ['$routeParams', '$scope', '$rootScope', '$http', function($routeParams, $scope, $rootScope, $http) {
	$scope.cur_repo = "";
	$scope.repos = [];
	$scope.packages = [];

	$scope.toggleRepo = function(repo) {
		$scope.cur_repo = repo;
	};

	getRepos = function() {
		return $http.get('/api/v1/acl/', {
			params: {}
		}).then(function(response) {
			$scope.repos = response.data.data;
		});
	};

	getRepos();
}])
.controller('AclInfoCtrl', ['$routeParams', '$scope', '$rootScope', '$http', function($routeParams, $scope, $rootScope, $http) {
	if ($routeParams.type != "groups" && $routeParams.type != "packages") {
		alert("Wrong type: " + $routeParams.type);
		return;
	}

	$scope.Name     = "";
	$scope.Repo     = "";
	$scope.Members  = [];
	$scope.Found    = false;
	$scope.NotFound = false;

	getACL = function() {
		return $http.get('/api/v1/acl/' + $routeParams.repo + '/' + $routeParams.type + '/' + $routeParams.name, {
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
.controller('AclSearchCtrl', ['$routeParams', '$scope', '$location', '$http', function($routeParams, $scope, $location, $http) {
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
;
