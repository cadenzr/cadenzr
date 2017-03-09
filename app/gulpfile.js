var gulp = require("gulp");
var browserify = require("browserify");
var source = require('vinyl-source-stream');
var buffer = require('vinyl-buffer');
var tsify = require("tsify");
var uglify = require('gulp-uglify');
var ts = require('gulp-typescript');
var vueify = require('vueify')
var tsProject = ts.createProject('./src/tsconfig.json');
var copy = require('gulp-copy');

var workingDir = 'working';

gulp.task('copy-vue-components', function() {
	return gulp
		.src('src/**/*.vue')
		.pipe(gulp.dest(workingDir + '/js'));
});

gulp.task('typescript', function() {
	var tsResult = tsProject.src()
	.pipe(tsProject());

	return tsResult.js.pipe(gulp.dest(workingDir + '/js'));
});

gulp.task('browserify', function() {
	    return browserify({
		            basedir: '.',
		            debug: true,
		            entries: [workingDir + '/js/main.js'],
		            cache: {},
		            packageCache: {}
		        })
		.transform(vueify)
	    .bundle()
	    .pipe(source('bundle.js'))
		.pipe(buffer())
		//.pipe(uglify())
	    .pipe(gulp.dest("dist"));
});

gulp.task('watch-ts', function () {
	gulp.watch('src/**/*.ts' , ['typescript']);
});

gulp.task('watch-js', function () {
	gulp.watch(workingDir + '/js/**/*.js' , ['browserify']);
});

gulp.task('watch-vue', function () {
	gulp.watch('src/**/*.vue' , ['copy-vue-components', 'browserify']);
});

gulp.task('watch', ['watch-ts', 'watch-js', 'watch-vue']);

gulp.task("default", function () {
	    return browserify({
		            basedir: '.',
		            debug: true,
		            entries: ['src/main.ts'],
		            cache: {},
		            packageCache: {}
		        })
	    .plugin(tsify, tsConfig.compilerOptions)
	    .bundle()
	    .pipe(source('bundle.js'))
		.pipe(buffer())
		//.pipe(uglify())
	    .pipe(gulp.dest("dist"));
	});
