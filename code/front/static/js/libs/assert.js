// Copyright (c) 2011 Jxck
//
// Originally from node.js (http://nodejs.org)
// Copyright Joyent, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the 'Software'), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
// ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

(function(a){function h(a,b){return void 0===b?""+b:"number"!=typeof b||!isNaN(b)&&isFinite(b)?"function"==typeof b||b instanceof RegExp?""+b:b:""+b}function i(a,b){return"string"==typeof a?b>a.length?a:a.slice(0,b):a}function j(a,b,c,d,e){throw new f.AssertionError({message:c,actual:a,expected:b,operator:d,stackStartFunction:e})}function k(a,b){a||j(a,!0,b,"==",f.ok)}function o(a,b){return a===b?!0:a instanceof Date&&b instanceof Date?a.getTime()===b.getTime():a instanceof RegExp&&b instanceof RegExp?a.source===b.source&&a.global===b.global&&a.multiline===b.multiline&&a.lastIndex===b.lastIndex&&a.ignoreCase===b.ignoreCase:"object"!=typeof a&&"object"!=typeof b?a==b:r(a,b)}function p(a){return null===a||void 0===a}function q(a){return"[object Arguments]"==Object.prototype.toString.call(a)}function r(a,b){if(p(a)||p(b))return!1;if(a.prototype!==b.prototype)return!1;if(q(a))return q(b)?(a=d.call(a),b=d.call(b),o(a,b)):!1;try{var g,h,c=e(a),f=e(b)}catch(i){return!1}if(c.length!=f.length)return!1;for(c.sort(),f.sort(),h=c.length-1;h>=0;h--)if(c[h]!=f[h])return!1;for(h=c.length-1;h>=0;h--)if(g=c[h],!o(a[g],b[g]))return!1;return!0}function v(a,b){return a&&b?"[object RegExp]"==Object.prototype.toString.call(b)?b.test(a):a instanceof b?!0:b.call({},a)===!0?!0:!1:!1}function w(a,b,c,d){var e;"string"==typeof c&&(d=c,c=null);try{b()}catch(f){e=f}if(d=(c&&c.name?" ("+c.name+").":".")+(d?" "+d:"."),a&&!e&&j(e,c,"Missing expected exception"+d),!a&&v(e,c)&&j(e,c,"Got unwanted exception"+d),a&&e&&c&&!v(e,c)||!a&&e)throw e}var b=Object.create||function(a){function b(){}if(!a)throw Error("no type");return b.prototype=a,new b},c={inherits:function(a,c){a.super_=c,a.prototype=b(c.prototype,{constructor:{value:a,enumerable:!1,writable:!0,configurable:!0}})}},d=Array.prototype.slice,e="function"==typeof Object.keys?Object.keys:function(a){var b=[];for(var c in a)b.push(c);return b},f=k;a.assert=f,"object"==typeof module&&"object"==typeof module.exports&&(module.exports=f),f.AssertionError=function(a){this.name="AssertionError",this.message=a.message,this.actual=a.actual,this.expected=a.expected,this.operator=a.operator;var b=a.stackStartFunction||j;Error.captureStackTrace&&Error.captureStackTrace(this,b)},c.inherits(f.AssertionError,Error),f.AssertionError.prototype.toString=function(){return this.message?[this.name+":",this.message].join(" "):[this.name+":",i(JSON.stringify(this.actual,h),128),this.operator,i(JSON.stringify(this.expected,h),128)].join(" ")},f.fail=j,f.ok=k,f.equal=function(a,b,c){a!=b&&j(a,b,c,"==",f.equal)},f.notEqual=function(a,b,c){a==b&&j(a,b,c,"!=",f.notEqual)},f.deepEqual=function(a,b,c){o(a,b)||j(a,b,c,"deepEqual",f.deepEqual)},f.notDeepEqual=function(a,b,c){o(a,b)&&j(a,b,c,"notDeepEqual",f.notDeepEqual)},f.strictEqual=function(a,b,c){a!==b&&j(a,b,c,"===",f.strictEqual)},f.notStrictEqual=function(a,b,c){a===b&&j(a,b,c,"!==",f.notStrictEqual)},f.throws=function(){w.apply(this,[!0].concat(d.call(arguments)))},f.doesNotThrow=function(){w.apply(this,[!1].concat(d.call(arguments)))},f.ifError=function(a){if(a)throw a},"function"==typeof define&&define.amd&&define("assert",function(){return f})})(this);
