% This is a sample Prolog facts file

parent(pam, bob).
parent(tom, bob).
parent(tom, liz).
parent(bob, ann).
parent(bob, pat).
parent(pat, jim).

male(tom).
male(bob).
male(jim).

female(pam).
female(liz).
female(ann).
female(pat).

grandparent(X, Y) :- parent(X, Z), parent(Z, Y).
