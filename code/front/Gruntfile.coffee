module.exports = (grunt) ->

	grunt.initConfig
		sourceDirectory: 'src'
		testDirectory: 'test'
		compiledHtmlDirectory: 'static/html'
		compiledCoffeeDirectory: 'static/js/src'
		compiledLessDirectory: 'static/css/src'

		shell:
			options:
				stdout: true
				stderr: true
				failOnError: true

			compileCoffee:
				command: 'iced --compile --runtime window --output <%= frontCoffeeCompiledDirectory %>/ <%= sourceDirectory %>/'

			compileHtml:
				command: 'find <%= sourceDirectory %> -name "*.html" | cpio -pdm <%= compiledHtmlDirectory %>'

			removeCompile:
				command: [
					'rm -rf <%= compiledHtmlDirectory %>',
					'rm -rf <%= compiledCoffeeDirectory %>',
					'rm -rf <%= compiledLessDirectory %>'
				].join ' && '

			test:
				command: 'echo "REIMPLEMENT THIS"'
				# command: 'karma start <%= testDirectory %>/karma.unit.conf.js --browsers PhantomJS --single-run'

		uglify:
			options:
				preserveComments: 'some'

			code:
				files: [
					expand: true
					cwd: '<%= compiledCoffeeDirectory %>/'
					src: ['**/*.js']
					dest: '<%= frontUglifiedDirectory %>/'
					ext: '.js'
				]

		less:
			development:
				files: [
					expand: true
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
					cwd: '<%= sourceDirectory %>/'
					src: ['**/*.less']
					dest: '<%= compiledLessDirectory %>/'
					ext: '.css'
				]

		watch:
			compile:
				files: ['<%= sourceDirectory %>/**/*.html', '<%= sourceDirectory %>/**/*.coffee', '<%= sourceDirectory %>/**/*.less']
				tasks: ['compile']

			test:
				files: ['<%= sourceDirectory %>/**/*.html', '<%= sourceDirectory %>/**/*.coffee', '<%= sourceDirectory %>/**/*.less']
				tasks: ['compile', 'test']

	grunt.loadNpmTasks 'grunt-contrib-uglify'
	grunt.loadNpmTasks 'grunt-contrib-less'
	grunt.loadNpmTasks 'grunt-contrib-watch'
	grunt.loadNpmTasks 'grunt-shell'

	grunt.registerTask 'default', ['compile']
	grunt.registerTask 'compile', ['shell:copyHtml', 'shell:removeCompile', 'shell:compileCoffee', 'less:development']
	grunt.registerTask 'compile-production', ['shell:copyHtml', 'shell:removeCompile', 'shell:compileCoffee', 'less:production']

	grunt.registerTask 'test', ['shell:test']

	grunt.registerTask 'make-ugly', ['shell:removeUglify', 'uglify']
	grunt.registerTask 'production', ['compile-production', 'make-ugly', 'shell:replaceCompiledWithUglified']
