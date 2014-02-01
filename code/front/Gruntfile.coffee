module.exports = (grunt) ->

	grunt.initConfig
		sourceDirectory: 'src'
		testDirectory: 'test'
		compiledHtmlDirectory: 'static/html'
		compiledCoffeeDirectory: 'static/js/src'
		compiledLessDirectory: 'static/css/src'
		uglifiedDirectory: 'uglified'

		shell:
			options:
				stdout: true
				stderr: true
				failOnError: true

			compileHtml:
				command: [
					'cd <%= sourceDirectory %>',
					'find . -name "*.html" | cpio -pmud ../<%= compiledHtmlDirectory %>'
				].join ' && '

			removeCompiled:
				command: [
					'rm -rf <%= compiledHtmlDirectory %>',
					'rm -rf <%= compiledCoffeeDirectory %>',
					'rm -rf <%= compiledLessDirectory %>'
				].join ' && '

			replaceCompiledWithUglified:
				command: [
					'rm -rf <%= compiledCoffeeDirectory %>',
					'mv <%= uglifiedDirectory %> <%= compiledCoffeeDirectory %>'
				].join ' && '

			removeUglified:
				command: 'rm -rf <%= uglifiedDirectory %>'

		uglify:
			options:
				preserveComments: 'some'

			code:
				files: [
					expand: true
					cwd: '<%= compiledCoffeeDirectory %>/'
					src: ['**/*.js']
					dest: '<%= uglifiedDirectory %>/'
					ext: '.js'
				]

		coffee:
			compile:
				files: [
					expand: true
					flatten: false
					cwd: '<%= sourceDirectory %>/'
					src: ['**/*.coffee']
					dest: '<%= compiledCoffeeDirectory %>/'
					ext: '.js'
				]

		less:
			development:
				files: [
					expand: true
					flatten: false
					cwd: '<%= sourceDirectory %>/'
					src: ['**/*.less']
					dest: '<%= compiledLessDirectory %>/'
					ext: '.css'
				]

			production:
				options:
					yuicompress: true
				files: [
					expand: true
					flatten: false
					cwd: '<%= sourceDirectory %>/'
					src: ['**/*.less']
					dest: '<%= compiledLessDirectory %>/'
					ext: '.css'
				]

		watch:
			compile:
				files: ['<%= sourceDirectory %>/**/*.html', '<%= sourceDirectory %>/**/*.coffee', '<%= sourceDirectory %>/**/*.less']
				tasks: ['compile']

	grunt.loadNpmTasks 'grunt-shell'
	grunt.loadNpmTasks 'grunt-contrib-uglify'
	grunt.loadNpmTasks 'grunt-contrib-less'
	grunt.loadNpmTasks 'grunt-contrib-watch'
	grunt.loadNpmTasks 'grunt-iced-coffee'

	grunt.registerTask 'default', ['compile']
	grunt.registerTask 'compile', ['shell:removeCompiled', 'shell:compileHtml', 'coffee:compile', 'less:development']
	grunt.registerTask 'compile-production', ['shell:removeCompiled', 'shell:compileHtml', 'coffee:compile', 'less:production']

	grunt.registerTask 'make-ugly', ['shell:removeUglified', 'uglify']
	grunt.registerTask 'production', ['compile-production', 'make-ugly', 'shell:replaceCompiledWithUglified', 'shell:removeUglified']
